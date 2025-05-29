package mysql

import (
	"context"

	"github.com/Xaytick/chat-zinx/chat-server/pkg/database"
	"github.com/Xaytick/chat-zinx/chat-server/pkg/model"
	"gorm.io/gorm"
)

// EnhancedUserDAO 增强的用户DAO，支持读写分离和分片
type EnhancedUserDAO struct {
	repo *database.Repository
}

// NewEnhancedUserDAO 创建新的增强用户DAO
func NewEnhancedUserDAO(repo *database.Repository) *EnhancedUserDAO {
	return &EnhancedUserDAO{
		repo: repo,
	}
}

// CreateUser 创建用户（写操作，使用主库）
func (dao *EnhancedUserDAO) CreateUser(ctx context.Context, user *model.User) error {
	opts := &database.QueryOptions{
		Operation: database.WriteOperation,
		ShardKey:  user.ID, // 使用用户ID作为分片键
		TableName: "users",
	}

	return dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Create(user).Error
	})
}

// GetUserByID 根据ID获取用户（读操作，使用从库）
func (dao *EnhancedUserDAO) GetUserByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User

	opts := &database.QueryOptions{
		Operation: database.ReadOperation,
		ShardKey:  userID,
		TableName: "users",
	}

	err := dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Where("id = ?", userID).First(&user).Error
	})

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUUID 根据UUID获取用户（读操作）
func (dao *EnhancedUserDAO) GetUserByUUID(ctx context.Context, uuid string) (*model.User, error) {
	var user model.User

	opts := &database.QueryOptions{
		Operation: database.ReadOperation,
		ShardKey:  uuid, // UUID也可以用作分片键
		TableName: "users",
	}

	err := dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Where("user_uuid = ?", uuid).First(&user).Error
	})

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户（读操作）
func (dao *EnhancedUserDAO) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	opts := &database.QueryOptions{
		Operation: database.ReadOperation,
		ShardKey:  username,
		TableName: "users",
	}

	err := dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Where("username = ?", username).First(&user).Error
	})

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息（写操作，使用主库）
func (dao *EnhancedUserDAO) UpdateUser(ctx context.Context, user *model.User) error {
	opts := &database.QueryOptions{
		Operation: database.WriteOperation,
		ShardKey:  user.ID,
		TableName: "users",
	}

	return dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Save(user).Error
	})
}

// DeleteUser 删除用户（写操作，使用主库）
func (dao *EnhancedUserDAO) DeleteUser(ctx context.Context, userID uint) error {
	opts := &database.QueryOptions{
		Operation: database.WriteOperation,
		ShardKey:  userID,
		TableName: "users",
	}

	return dao.repo.Execute(ctx, opts, func(db *gorm.DB) error {
		return db.Delete(&model.User{}, userID).Error
	})
}

// BatchCreateUsers 批量创建用户
func (dao *EnhancedUserDAO) BatchCreateUsers(ctx context.Context, users []*model.User) error {
	items := make([]interface{}, len(users))
	for i, user := range users {
		items[i] = user
	}

	return dao.repo.BatchInsert(ctx, "users", items, func(item interface{}) interface{} {
		user := item.(*model.User)
		return user.ID
	})
}

// SearchUsers 搜索用户（跨分片查询）
func (dao *EnhancedUserDAO) SearchUsers(ctx context.Context, keyword string, limit int) ([]*model.User, error) {
	results, err := dao.repo.CrossShardQuery(ctx, "users", func(db *gorm.DB) *gorm.DB {
		return db.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%").Limit(limit)
	})

	if err != nil {
		return nil, err
	}

	users := make([]*model.User, 0, len(results))
	for _, result := range results {
		user := &model.User{}
		// 这里需要手动映射，或者使用更好的序列化方法
		if id, ok := result["id"].(uint); ok {
			user.ID = id
		}
		if username, ok := result["username"].(string); ok {
			user.Username = username
		}
		if email, ok := result["email"].(string); ok {
			user.Email = email
		}
		// ... 映射其他字段
		users = append(users, user)
	}

	return users, nil
}

// GetUserStats 获取用户统计信息（跨分片聚合）
func (dao *EnhancedUserDAO) GetUserStats(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	results, err := dao.repo.CrossShardQuery(ctx, "users", func(db *gorm.DB) *gorm.DB {
		return db.Select("COUNT(*) as total")
	})

	if err != nil {
		return nil, err
	}

	var total int64
	for _, result := range results {
		if count, ok := result["total"].(int64); ok {
			total += count
		}
	}

	stats["total_users"] = total
	return stats, nil
}
