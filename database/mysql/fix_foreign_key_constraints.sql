-- ========================================
-- 修复User Service外键约束问题
-- ========================================

USE jobfirst;

-- 1. 清理resume_v3表中的无效template_id引用
UPDATE resume_v3 
SET template_id = NULL 
WHERE template_id IS NOT NULL 
AND template_id NOT IN (SELECT id FROM resume_templates);

-- 2. 清理resume_skills表中的无效引用
DELETE FROM resume_skills 
WHERE resume_id NOT IN (SELECT id FROM resume_v3);

DELETE FROM resume_skills 
WHERE skill_id NOT IN (SELECT id FROM skills);

-- 3. 显示修复结果
SELECT 'Foreign key constraints fixed!' as status;
SELECT COUNT(*) as resume_v3_count FROM resume_v3;
SELECT COUNT(*) as resume_skills_count FROM resume_skills;
SELECT COUNT(*) as resume_templates_count FROM resume_templates;
