package models

import "time"

const (
	QuestionStatusPending  = "pending"
	QuestionStatusAnswered = "answered"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	BasaltID  string    `gorm:"uniqueIndex;not null" json:"basalt_id"`
	Email     string    `gorm:"index;not null" json:"email"`
	Name      string    `gorm:"size:120;not null" json:"name"`
	Slug      string    `gorm:"uniqueIndex;size:80;not null" json:"slug"`
	BoxTitle  string    `gorm:"size:120;not null" json:"box_title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Questions []Question `json:"questions,omitempty" gorm:"foreignKey:OwnerID;constraint:OnDelete:CASCADE;"`
}

type Question struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	OwnerID         uint       `gorm:"index;not null" json:"owner_id"`
	AnonymousName   string     `gorm:"size:40;not null;default:'Anonymous'" json:"anonymous_name"`
	Content         string     `gorm:"type:text;not null" json:"content"`
	Answer          string     `gorm:"type:text" json:"answer"`
	Status          string     `gorm:"size:20;index;not null;default:'pending'" json:"status"`
	BackgroundColor string     `gorm:"size:16;not null;default:'#fff4d6'" json:"background_color"`
	AnsweredAt      *time.Time `json:"answered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
