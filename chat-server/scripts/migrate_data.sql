-- =========================================
-- 数据迁移脚本: 从单库迁移到读写分离+分片架构
-- =========================================

-- 步骤1: 导出现有数据
-- 执行命令: mysqldump -u root -p chat_app > backup_chat_app.sql

-- 步骤2: 连接到现有数据库，查看数据分布
USE chat_app;

-- 查看现有数据统计
SELECT 'users' as table_name, COUNT(*) as record_count FROM users
UNION ALL
SELECT 'groups' as table_name, COUNT(*) as record_count FROM groups  
UNION ALL
SELECT 'group_members' as table_name, COUNT(*) as record_count FROM group_members
UNION ALL
SELECT 'group_messages' as table_name, COUNT(*) as record_count FROM group_messages;

-- 步骤3: 分析用户数据分布 (用于分片)
SELECT 
    CASE 
        WHEN id % 2 = 0 THEN 'shard_00'
        ELSE 'shard_01'
    END as shard,
    COUNT(*) as user_count
FROM users
GROUP BY id % 2;

-- 步骤4: 分析群组消息分布 (用于分片)
SELECT 
    CASE 
        WHEN group_id % 2 = 0 THEN 'shard_00'
        ELSE 'shard_01'
    END as shard,
    COUNT(*) as message_count
FROM group_messages
GROUP BY group_id % 2;

-- =========================================
-- 实际迁移脚本 (在新环境中执行)
-- =========================================

-- 步骤5: 迁移用户数据到分片
-- 分片0 (偶数ID)
/*
INSERT INTO chat_app_shard_00.users_00 
SELECT * FROM chat_app.users WHERE id % 2 = 0;
*/

-- 分片1 (奇数ID)  
/*
INSERT INTO chat_app_shard_01.users_01
SELECT * FROM chat_app.users WHERE id % 2 = 1;
*/

-- 步骤6: 迁移群组消息到分片
-- 分片0 (偶数group_id)
/*
INSERT INTO chat_app_shard_00.group_messages_00
SELECT * FROM chat_app.group_messages WHERE group_id % 2 = 0;
*/

-- 分片1 (奇数group_id)
/*
INSERT INTO chat_app_shard_01.group_messages_01  
SELECT * FROM chat_app.group_messages WHERE group_id % 2 = 1;
*/

-- =========================================
-- 验证迁移结果
-- =========================================

-- 验证用户数据迁移
/*
SELECT 'original' as source, COUNT(*) as count FROM chat_app.users
UNION ALL
SELECT 'shard_00' as source, COUNT(*) as count FROM chat_app_shard_00.users_00
UNION ALL  
SELECT 'shard_01' as source, COUNT(*) as count FROM chat_app_shard_01.users_01;
*/

-- 验证群组消息迁移
/*
SELECT 'original' as source, COUNT(*) as count FROM chat_app.group_messages
UNION ALL
SELECT 'shard_00' as source, COUNT(*) as count FROM chat_app_shard_00.group_messages_00
UNION ALL
SELECT 'shard_01' as source, COUNT(*) as count FROM chat_app_shard_01.group_messages_01;
*/ 