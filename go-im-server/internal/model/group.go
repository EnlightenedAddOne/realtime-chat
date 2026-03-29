package model

import "time"

const (
	GroupRoleMember = 1
	GroupRoleAdmin  = 2
	GroupRoleOwner  = 3

	GroupRequestStatusPending  = 0
	GroupRequestStatusAccepted = 1
	GroupRequestStatusRejected = 2

	GroupRequestTypeJoinApply = 1
	GroupRequestTypeInvite    = 2
)

// Group 群组表
type Group struct {
	ID                    uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OwnerID               uint      `gorm:"index;not null" json:"owner_id"`
	Name                  string    `gorm:"type:varchar(50);not null" json:"name"`
	Avatar                string    `gorm:"type:varchar(255)" json:"avatar"`
	Announcement          string    `gorm:"type:text" json:"announcement"`
	AnnouncementUpdatedAt time.Time `json:"announcement_updated_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (Group) TableName() string {
	return "groups"
}

// GroupMember 群组成员表
type GroupMember struct {
	GroupID  uint      `gorm:"primaryKey;index;not null" json:"group_id"`
	UserID   uint      `gorm:"primaryKey;index;not null" json:"user_id"`
	Role     int8      `gorm:"type:smallint;default:1;comment:'1:Member, 2:Admin, 3:Owner'" json:"role"`
	Nickname string    `gorm:"type:varchar(50)" json:"nickname"`
	JoinedAt time.Time `json:"joined_at"`
}

// GroupMemberDetail is the group member DTO used by member list API.
type GroupMemberDetail struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	Role      int8      `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

func (GroupMember) TableName() string {
	return "group_members"
}

// GroupRequest 群组入群申请表
type GroupRequest struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	GroupID   uint      `gorm:"index;not null" json:"group_id"`
	InviterID *uint     `gorm:"index" json:"inviter_id,omitempty"`
	Remark    string    `gorm:"type:text" json:"remark"`
	Status    int       `gorm:"type:smallint;default:0;comment:'0:Pending, 1:Accepted, 2:Rejected'" json:"status"`
	Type      int       `gorm:"type:smallint;default:1;index;comment:'1:JoinApply, 2:Invite'" json:"type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Inviter   *User     `gorm:"foreignKey:InviterID" json:"inviter,omitempty"`
	Group     *Group    `gorm:"foreignKey:GroupID" json:"group,omitempty"`
}

func (GroupRequest) TableName() string {
	return "group_requests"
}
