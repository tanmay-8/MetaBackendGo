package db

import (
	"backend/src/models"
	"context"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbAdapter struct {
	Db *mongo.Database
}

func NewDbAdapter(ctx context.Context) (*DbAdapter, error) {
	uri := os.Getenv("BACKEND_MONGO_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		slog.Error("Error connecting to mongo", err)
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		slog.Error("Error pinging mongo", err)
		return nil, err
	}
	db := client.Database("metamorphosis")
	return &DbAdapter{Db: db}, nil
}

func (d DbAdapter) Close(ctx context.Context) error {
	return d.Db.Client().Disconnect(ctx)
}

// Participant Operations
func (d DbAdapter) GetNextPID(ctx context.Context) (int, error) {
	var counter models.Counter
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := d.Db.Collection("counters").FindOneAndUpdate(
		ctx,
		bson.M{"name": "participant_pid"},
		bson.M{"$inc": bson.M{"seq": 1}},
		opts,
	).Decode(&counter)
	if err != nil {
		return 0, err
	}
	return counter.Seq, nil
}

func (d DbAdapter) CreateParticipant(ctx context.Context, participant models.Participant) (int, error) {
	pid, err := d.GetNextPID(ctx)
	if err != nil {
		return 0, err
	}
	participant.PID = pid
	participant.CreatedAt = time.Now()
	participant.UpdatedAt = time.Now()
	_, err = d.Db.Collection("participants").InsertOne(ctx, participant)
	return pid, err
}

func (d DbAdapter) GetParticipant(ctx context.Context, pid int) (models.Participant, error) {
	var participant models.Participant
	err := d.Db.Collection("participants").FindOne(ctx, bson.M{"pid": pid}).Decode(&participant)
	return participant, err
}

// Registration Operations
func (d DbAdapter) CreateRegistration(ctx context.Context, reg models.Registration) (string, error) {
	reg.CreatedAt = time.Now()
	reg.UpdatedAt = time.Now()
	result, err := d.Db.Collection("registrations").InsertOne(ctx, reg)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (d DbAdapter) GetRegistration(ctx context.Context, id string) (models.Registration, error) {
	var reg models.Registration
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return reg, err
	}
	err = d.Db.Collection("registrations").FindOne(ctx, bson.M{"_id": objID}).Decode(&reg)
	return reg, err
}
