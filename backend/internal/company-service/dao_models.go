package main

import (
	"time"
)

// CompanyDAO 企业DAO数据模型
type CompanyDAO struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	CompanyID         uint      `json:"company_id" gorm:"not null"`
	Name              string    `json:"name" gorm:"size:200;not null"`
	Description       string    `json:"description" gorm:"type:text"`
	GovernanceToken   string    `json:"governance_token" gorm:"size:100"`
	ContractAddress   string    `json:"contract_address" gorm:"size:100"`
	TotalSupply       uint64    `json:"total_supply" gorm:"default:0"`
	CirculatingSupply uint64    `json:"circulating_supply" gorm:"default:0"`
	VotingThreshold   float64   `json:"voting_threshold" gorm:"type:decimal(5,2);default:50.00"`
	ProposalThreshold uint64    `json:"proposal_threshold" gorm:"default:1000"`
	VotingPeriod      int       `json:"voting_period" gorm:"default:7"`
	ExecutionDelay    int       `json:"execution_delay" gorm:"default:1"`
	Status            string    `json:"status" gorm:"size:20;default:active"`
	CreatedBy         uint      `json:"created_by" gorm:"not null"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	
	// 关联数据
	Company   Company     `json:"company" gorm:"foreignKey:CompanyID"`
	Members   []DAOMember `json:"members" gorm:"foreignKey:DAOID"`
	Proposals []DAOProposal `json:"proposals" gorm:"foreignKey:DAOID"`
	Teams     []AutonomousTeam `json:"teams" gorm:"foreignKey:DAOID"`
}

// DAOMember DAO成员数据模型
type DAOMember struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	DAOID            uint      `json:"dao_id" gorm:"not null"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	Role             string    `json:"role" gorm:"size:50;default:member"`
	VotingPower       uint64    `json:"voting_power" gorm:"default:0"`
	TokenBalance      uint64    `json:"token_balance" gorm:"default:0"`
	ContributionScore int       `json:"contribution_score" gorm:"default:0"`
	JoinedAt         time.Time `json:"joined_at"`
	Status           string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// 关联数据
	DAO  CompanyDAO `json:"dao" gorm:"foreignKey:DAOID"`
	User User       `json:"user" gorm:"foreignKey:UserID"`
}

// DAOProposal DAO提案数据模型
type DAOProposal struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	DAOID         uint      `json:"dao_id" gorm:"not null"`
	Title         string    `json:"title" gorm:"size:200;not null"`
	Description   string    `json:"description" gorm:"type:text"`
	ProposerID    uint      `json:"proposer_id" gorm:"not null"`
	ProposalType  string    `json:"proposal_type" gorm:"size:50"`
	ProposalData  string    `json:"proposal_data" gorm:"type:json"`
	VotesFor      uint64    `json:"votes_for" gorm:"default:0"`
	VotesAgainst  uint64    `json:"votes_against" gorm:"default:0"`
	TotalVotes    uint64    `json:"total_votes" gorm:"default:0"`
	VotingThreshold uint64  `json:"voting_threshold" gorm:"default:0"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
	ExecutionTime *time.Time `json:"execution_time"`
	Status        string    `json:"status" gorm:"size:20;default:draft"`
	ExecutionResult string  `json:"execution_result" gorm:"type:json"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	// 关联数据
	DAO      CompanyDAO `json:"dao" gorm:"foreignKey:DAOID"`
	Proposer User      `json:"proposer" gorm:"foreignKey:ProposerID"`
	Votes    []DAOVote `json:"votes" gorm:"foreignKey:ProposalID"`
}

// DAOVote DAO投票记录数据模型
type DAOVote struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProposalID  uint      `json:"proposal_id" gorm:"not null"`
	VoterID     uint      `json:"voter_id" gorm:"not null"`
	VoteType    string    `json:"vote_type" gorm:"size:20;not null"`
	VotingPower uint64    `json:"voting_power" gorm:"not null"`
	VoteReason  string    `json:"vote_reason" gorm:"type:text"`
	VotedAt     time.Time `json:"voted_at"`
	CreatedAt   time.Time `json:"created_at"`
	
	// 关联数据
	Proposal DAOProposal `json:"proposal" gorm:"foreignKey:ProposalID"`
	Voter    User        `json:"voter" gorm:"foreignKey:VoterID"`
}

// AutonomousTeam 自主管理团队数据模型
type AutonomousTeam struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	DAOID          uint      `json:"dao_id" gorm:"not null"`
	Name           string    `json:"name" gorm:"size:200;not null"`
	Description    string    `json:"description" gorm:"type:text"`
	LeaderID       uint      `json:"leader_id" gorm:"not null"`
	Budget         uint64    `json:"budget" gorm:"default:0"`
	MaxMembers     int       `json:"max_members" gorm:"default:50"`
	CurrentMembers int       `json:"current_members" gorm:"default:0"`
	Status         string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// 关联数据
	DAO     CompanyDAO    `json:"dao" gorm:"foreignKey:DAOID"`
	Leader  User          `json:"leader" gorm:"foreignKey:LeaderID"`
	Members []TeamMember `json:"members" gorm:"foreignKey:TeamID"`
}

// TeamMember 团队成员数据模型
type TeamMember struct {
	ID               uint      `json:"id" gorm:"primaryKey"`
	TeamID           uint      `json:"team_id" gorm:"not null"`
	UserID           uint      `json:"user_id" gorm:"not null"`
	Role             string    `json:"role" gorm:"size:50;default:member"`
	ContributionScore int      `json:"contribution_score" gorm:"default:0"`
	JoinedAt         time.Time `json:"joined_at"`
	Status           string    `json:"status" gorm:"size:20;default:active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	
	// 关联数据
	Team AutonomousTeam `json:"team" gorm:"foreignKey:TeamID"`
	User User           `json:"user" gorm:"foreignKey:UserID"`
}

// DAOActivity DAO活动记录数据模型
type DAOActivity struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	DAOID        uint      `json:"dao_id" gorm:"not null"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	ActivityType  string    `json:"activity_type" gorm:"size:50;not null"`
	ActivityData string    `json:"activity_data" gorm:"type:json"`
	Description  string    `json:"description" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	
	// 关联数据
	DAO  CompanyDAO `json:"dao" gorm:"foreignKey:DAOID"`
	User User       `json:"user" gorm:"foreignKey:UserID"`
}

// DAOSetting DAO配置数据模型
type DAOSetting struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	DAOID       uint      `json:"dao_id" gorm:"not null"`
	SettingKey  string    `json:"setting_key" gorm:"size:100;not null"`
	SettingValue string   `json:"setting_value" gorm:"type:text"`
	SettingType string    `json:"setting_type" gorm:"size:20;default:string"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// 关联数据
	DAO CompanyDAO `json:"dao" gorm:"foreignKey:DAOID"`
}

// 提案类型枚举
const (
	ProposalTypeBudgetAllocation = "budget_allocation"
	ProposalTypeTeamStructure    = "team_structure"
	ProposalTypeProjectApproval  = "project_approval"
	ProposalTypePolicyChange     = "policy_change"
	ProposalTypeMemberManagement = "member_management"
	ProposalTypeOther            = "other"
)

// 投票类型枚举
const (
	VoteTypeFor     = "for"
	VoteTypeAgainst = "against"
	VoteTypeAbstain = "abstain"
)

// 提案状态枚举
const (
	ProposalStatusDraft     = "draft"
	ProposalStatusActive    = "active"
	ProposalStatusPassed    = "passed"
	ProposalStatusRejected  = "rejected"
	ProposalStatusExecuted  = "executed"
	ProposalStatusExpired   = "expired"
)

// 成员角色枚举
const (
	MemberRoleFounder     = "founder"
	MemberRoleAdmin       = "admin"
	MemberRoleMember      = "member"
	MemberRoleContributor = "contributor"
)

// 团队角色枚举
const (
	TeamRoleLeader      = "leader"
	TeamRoleAdmin       = "admin"
	TeamRoleMember      = "member"
	TeamRoleContributor = "contributor"
)

// 活动类型枚举
const (
	ActivityTypeProposalCreated  = "proposal_created"
	ActivityTypeProposalVoted    = "proposal_voted"
	ActivityTypeProposalExecuted = "proposal_executed"
	ActivityTypeMemberJoined     = "member_joined"
	ActivityTypeMemberLeft       = "member_left"
	ActivityTypeTeamCreated      = "team_created"
	ActivityTypeBudgetAllocated  = "budget_allocated"
	ActivityTypeOther            = "other"
)
