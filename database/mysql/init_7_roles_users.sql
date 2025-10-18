-- JobFirst 7个角色用户初始化脚本
-- 用于创建完整的7个角色测试用户
-- 密码统一使用: testuser123 (bcrypt加密)

USE jobfirst;

-- 插入7个角色的测试用户
INSERT IGNORE INTO users (username, email, password_hash, role, status) VALUES
-- 1. super_admin (已存在)
('admin', 'admin@jobfirst.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'super_admin', 'active'),

-- 2. system_admin
('testuser2', 'test2@example.com', '$2a$10$6GUURUtL7QjG2DrPefuoauqDoCHhB0ybMptsMhNcdMZBOrX1tDRNq', 'system_admin', 'active'),

-- 3. dev_lead
('testuser3', 'testuser3@example.com', '$2a$10$2agLNzGvkA154hu2ybX85eYG.3uyQSmXdnH7FuqA3TGF2gdkSOcmi', 'dev_lead', 'active'),

-- 4. frontend_dev
('testuser4', 'testuser4@example.com', '$2a$10$W5yGlt64M9uFLb/rcubm2Ok0NKy2pbwLsnwYgxlZkZi8Ht2YouvkK', 'frontend_dev', 'active'),

-- 5. backend_dev
('testuser5', 'testuser5@example.com', '$2a$10$B9BU5fhl7O8tfG.xxX/7o.LM/VWcaCB2HPAX9HSIRmzSPj3ipsIiG', 'backend_dev', 'active'),

-- 6. qa_engineer
('testuser6', 'testuser6@example.com', '$2a$10$HZYFXoVjzvQHQZgHFFxTGOw8BOekgJyfJbuRtvP3PlsQF3MCBsIie', 'qa_engineer', 'active'),

-- 7. guest (已存在)
('szjason72', '347399@qq.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'guest', 'active'),

-- 8. 额外的guest用户
('testuser', 'test@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'guest', 'active'),

-- 9. 另一个guest用户
('testuser7', 'testuser7@example.com', '$2a$10$VwsFjIqCPlg0Xwahqm0OI.fS4mwCH1T/WDcqh.rQMYvh.AVUxGU4i', 'guest', 'active');

-- 显示插入结果
SELECT 
    id,
    username,
    email,
    role,
    status,
    created_at
FROM users 
WHERE username IN ('admin', 'testuser2', 'testuser3', 'testuser4', 'testuser5', 'testuser6', 'szjason72', 'testuser', 'testuser7')
ORDER BY role, id;

-- 显示角色统计
SELECT 
    role,
    COUNT(*) as user_count
FROM users 
WHERE username IN ('admin', 'testuser2', 'testuser3', 'testuser4', 'testuser5', 'testuser6', 'szjason72', 'testuser', 'testuser7')
GROUP BY role
ORDER BY role;
