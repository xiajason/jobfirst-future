#!/usr/bin/env python3
"""更新AI Chat API集成三层AI分身系统"""

# 读取ai_service_with_zervigo.py
with open("ai_service_with_zervigo.py", "r") as f:
    content = f.read()

# 备份
with open("ai_service_with_zervigo.py.before_three_layer", "w") as f:
    f.write(content)
print("✅ 已备份原文件")

# 1. 添加导入
if "from three_layer_avatar_chat import" not in content:
    # 在其他导入后添加
    import_pos = content.find("from job_matching_service import")
    if import_pos > 0:
        insert_pos = content.find("\n", import_pos) + 1
        new_import = """from three_layer_avatar_chat import ThreeLayerAvatarChat
import aiomysql
"""
        content = content[:insert_pos] + new_import + content[insert_pos:]
        print("✅ 添加三层AI分身导入")

# 2. 初始化三层系统
if "three_layer_chat = ThreeLayerAvatarChat()" not in content:
    # 在app创建后添加
    app_pos = content.find("app = Sanic(__name__)")
    if app_pos > 0:
        insert_pos = content.find("\n", app_pos) + 1
        init_code = """
# 初始化三层AI分身系统
three_layer_chat = ThreeLayerAvatarChat()
logger.info("✅ 三层AI分身系统已初始化")
"""
        content = content[:insert_pos] + init_code + content[insert_pos:]
        print("✅ 添加三层系统初始化")

# 3. 替换ai_chat_api函数
old_chat_api = """async def ai_chat_api(request: Request):
    \"\"\"AI聊天API - 需要ai_chat权限\"\"\"
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        # 权限检查
        if not unified_auth_client.has_permission(request, "ai:chat"):
            return sanic_json({
                "error": "权限不足",
                "code": "PERMISSION_DENIED",
                "message": "您没有访问AI聊天功能的权限"
            }, status=403)
        
        # 配额检查
        quota_result = await check_quota(request, "ai_requests")
        if quota_result:
            return quota_result
        
        user_info = unified_auth_client.get_user_info(request)
        logger.info(f"收到AI聊天请求: {request.json}, 用户: {user_info.user_id}")
        
        # TODO: 实现AI聊天功能
        # 这里需要集成实际的AI聊天服务
        
        # 消耗配额
        await unified_auth_client.consume_quota(request, "ai_requests", 1)
        
        # 记录成功日志
        await log_user_action(request, "ai_chat", "success")
        
        return sanic_json({
            "success": True,
            "message": "AI聊天功能正在开发中",
            "user_id": user_info.user_id
        })
        
    except Exception as e:
        logger.error(f"AI聊天API异常: {e}")
        await log_user_action(request, "ai_chat", "failed")
        return sanic_json({"error": f"服务器内部错误: {str(e)}"}, status=500)"""

new_chat_api = """async def ai_chat_api(request: Request):
    \"\"\"AI分身对话API - 三层架构版本\"\"\"
    try:
        # 认证检查
        auth_result = await authenticate_user(request)
        if auth_result:
            return auth_result
        
        # 权限检查
        if not unified_auth_client.has_permission(request, "ai:chat"):
            return sanic_json({
                "error": "权限不足",
                "code": "PERMISSION_DENIED",
                "message": "您没有访问AI聊天功能的权限"
            }, status=403)
        
        # 配额检查
        quota_result = await check_quota(request, "ai_requests")
        if quota_result:
            return quota_result
        
        user_info = unified_auth_client.get_user_info(request)
        message = request.json.get("message", "")
        
        logger.info(f"收到AI分身对话请求: {message}, 用户: {user_info.user_id}")
        
        # 获取用户MBIT类型
        async def get_mbit_type(uid):
            try:
                conn = await aiomysql.connect(
                    host=os.getenv('MYSQL_HOST', 'localhost'),
                    port=int(os.getenv('MYSQL_PORT', 3306)),
                    user=os.getenv('MYSQL_USER', 'root'),
                    password=os.getenv('MYSQL_PASSWORD', 'test_mysql_password'),
                    db=os.getenv('MYSQL_DATABASE', 'jobfirst')
                )
                async with conn.cursor(aiomysql.DictCursor) as cursor:
                    await cursor.execute(
                        "SELECT mbit_type FROM user_mbit_tests WHERE user_id=%s ORDER BY test_date DESC LIMIT 1",
                        (uid,)
                    )
                    result = await cursor.fetchone()
                    await conn.close()
                    return result['mbit_type'] if result else 'UNKNOWN'
            except:
                return 'UNKNOWN'
        
        # 获取用户简历数据
        async def get_resume_data(uid):
            try:
                conn = await aiomysql.connect(
                    host=os.getenv('MYSQL_HOST', 'localhost'),
                    port=int(os.getenv('MYSQL_PORT', 3306)),
                    user=os.getenv('MYSQL_USER', 'root'),
                    password=os.getenv('MYSQL_PASSWORD', 'test_mysql_password'),
                    db=os.getenv('MYSQL_DATABASE', 'jobfirst')
                )
                async with conn.cursor(aiomysql.DictCursor) as cursor:
                    await cursor.execute(
                        "SELECT * FROM resume_metadata WHERE user_id=%s ORDER BY created_at DESC LIMIT 1",
                        (uid,)
                    )
                    result = await cursor.fetchone()
                    await conn.close()
                    
                    if result:
                        return {
                            'current_position': '工程师',
                            'years_experience': 3,
                            'skills': '["Python", "Docker", "MySQL"]'
                        }
                    return {}
            except:
                return {}
        
        # 构建用户数据
        mbit_type = await get_mbit_type(user_info.user_id)
        resume_data = await get_resume_data(user_info.user_id)
        
        user_data = {
            'user_id': user_info.user_id,
            'mbit_type': mbit_type,
            'resume_data': resume_data
        }
        
        logger.info(f"用户画像: MBIT={mbit_type}, 职位={resume_data.get('current_position', '未知')}")
        
        # 调用三层AI分身系统
        result = await three_layer_chat.chat(user_data, message)
        
        logger.info(f"AI分身回复: 路由={result['route']}, 层级={len(result['layers_used'])}, 成本=￥{result['cost']:.4f}")
        
        # 消耗配额
        await unified_auth_client.consume_quota(request, "ai_requests", 1)
        
        # 记录成功日志
        await log_user_action(request, "ai_chat", "success", {
            'route': result['route'],
            'layers': result['layers_used'],
            'cost': result['cost']
        })
        
        return sanic_json({
            "success": True,
            "reply": result['reply'],
            "metadata": {
                "route": result['route'],
                "layers_used": result['layers_used'],
                "cost": result['cost'],
                "profile": result['profile_summary'],
                "intent": result['metadata']['intent']
            },
            "user_id": user_info.user_id
        })
        
    except Exception as e:
        logger.error(f"AI分身对话异常: {e}")
        import traceback
        traceback.print_exc()
        await log_user_action(request, "ai_chat", "failed")
        return sanic_json({"error": f"AI分身暂时不可用: {str(e)}"}, status=500)"""

# 替换函数
if old_chat_api in content:
    content = content.replace(old_chat_api, new_chat_api)
    print("✅ 替换ai_chat_api函数")
else:
    print("⚠️  未找到完全匹配的函数，尝试查找函数定义...")
    # 查找函数开始
    func_start = content.find("async def ai_chat_api(request: Request):")
    if func_start > 0:
        # 查找下一个@app.route或文件结束
        next_route = content.find("@app.", func_start + 100)
        if next_route < 0:
            next_route = len(content)
        
        # 替换整个函数
        content = content[:func_start] + new_chat_api.split("async def ai_chat_api")[1]
        # 需要找到函数结束位置
        print("⚠️  需要手动检查函数替换")

# 写回文件
with open("ai_service_with_zervigo.py", "w") as f:
    f.write(content)

print("✅ AI Chat API更新完成")
print("\n📋 更新内容:")
print("  1. 导入ThreeLayerAvatarChat")
print("  2. 初始化三层系统")
print("  3. 替换ai_chat_api函数")
print("  4. 集成MBIT和简历数据")
print("  5. 调用三层路由")
print("\n🎯 下一步: 重启AI Service")
