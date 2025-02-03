package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Counter model
type Counter struct {
	ID  primitive.ObjectID `bson:"_id,omitempty"`
	Seq int                `bson:"seq" json:"seq"`
}

// Participant model
type Participant struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PID         int                `bson:"pid" json:"pid"` // Unique participant ID
	Name        string             `bson:"name" json:"name"`
	Email       string             `bson:"email" json:"email"`
	Phone       string             `bson:"phone" json:"phone"`
	CollegeName string             `bson:"collegeName" json:"collegeName"`
	YearOfStudy int                `bson:"yearOfStudy" json:"yearOfStudy"`
	DualBoot    bool               `bson:"dualBoot" json:"dualBoot"`
	MailSent    bool               `bson:"mailSent,omitempty" json:"mailSent"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

// Registration model
type Registration struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	NumOfParticipants int                `bson:"numOfParticipants" json:"numOfParticipants"`
	Participants      []int              `bson:"participants" json:"participants"` // Storing participant IDs
	TotalAmount       int                `bson:"totalAmount" json:"totalAmount"`
	TransactionID     string             `bson:"transactionId" json:"transactionId"`
	TransactionImage  string             `bson:"transactionImage" json:"transactionImage"`
	MailSent          bool               `bson:"mailSent,omitempty" json:"mailSent"`
	CreatedAt         time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt         time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
