package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MongoURI = "mongodb://localhost:27017"

// InitMongo initializes a MongoDB connection
func InitMongo(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		log.Fatal("Mongo connect:", err)
	}

	return client, nil
}

// SetupCoordinatesCollection creates the coordinates collection with validation
func SetupCoordinatesCollection(ctx context.Context, db *mongo.Database) error {
	// Check if collection already exists
	collectionName := "coordinates"
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, coll := range collections {
		if coll == collectionName {
			log.Printf("Collection '%s' already exists", collectionName)
			return nil
		}
	}

	// Define validation schema for coordinates collection
	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"user_id", "latitude", "longitude", "timestamp"},
			"properties": bson.M{
				"user_id": bson.M{
					"bsonType":    "string",
					"description": "ID of the user who generated this coordinate",
				},
				"latitude": bson.M{
					"bsonType":    "double",
					"description": "Latitude coordinate",
				},
				"longitude": bson.M{
					"bsonType":    "double",
					"description": "Longitude coordinate",
				},
				"altitude": bson.M{
					"bsonType":    "double",
					"description": "Altitude in meters (optional)",
				},
				"accuracy": bson.M{
					"bsonType":    "double",
					"description": "Accuracy of the coordinates in meters (optional)",
				},
				"speed": bson.M{
					"bsonType":    "double",
					"description": "Speed in meters per second (optional)",
				},
				"timestamp": bson.M{
					"bsonType":    "date",
					"description": "Time when the coordinate was recorded",
				},
				"activity_type": bson.M{
					"bsonType":    "string",
					"description": "Type of activity (e.g., running, cycling, walking)",
				},
			},
		},
	}

	// Create collection with validation
	opts := options.CreateCollection().SetValidator(validator)
	if err := db.CreateCollection(ctx, collectionName, opts); err != nil {
		return err
	}

	// Create indexes for better query performance
	coll := db.Collection(collectionName)

	// Index on user_id for filtering by user
	userIdIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}

	// Index on timestamp for time-based queries
	timestampIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "timestamp", Value: 1}},
	}

	// Compound index on user_id and timestamp for efficient user timeline queries
	userTimeIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "timestamp", Value: 1}},
	}

	// Geospatial index for location-based queries
	geoIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}

	// Create all indexes
	_, err = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		userIdIndex,
		timestampIndex,
		userTimeIndex,
		geoIndex,
	})

	if err != nil {
		return err
	}

	log.Printf("Collection '%s' created successfully with indexes", collectionName)
	return nil
}

// SetupLocationsCollection creates the locations collection with validation
func SetupLocationsCollection(ctx context.Context, db *mongo.Database) error {
	// Check if collection already exists
	collectionName := "locations"
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, coll := range collections {
		if coll == collectionName {
			log.Printf("Collection '%s' already exists", collectionName)
			return nil
		}
	}

	// Define validation schema for locations collection
	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"name", "coordinates", "created_at"},
			"properties": bson.M{
				"name": bson.M{
					"bsonType":    "string",
					"description": "Name of the location",
				},
				"description": bson.M{
					"bsonType":    "string",
					"description": "Description of the location",
				},
				"coordinates": bson.M{
					"bsonType": "object",
					"required": []string{"latitude", "longitude"},
					"properties": bson.M{
						"latitude": bson.M{
							"bsonType":    "double",
							"description": "Latitude coordinate",
						},
						"longitude": bson.M{
							"bsonType":    "double",
							"description": "Longitude coordinate",
						},
					},
				},
				"address": bson.M{
					"bsonType":    "string",
					"description": "Physical address of the location",
				},
				"category": bson.M{
					"bsonType":    "string",
					"description": "Category of the location (e.g., restaurant, park, museum)",
				},
				"tags": bson.M{
					"bsonType":    "array",
					"description": "Tags associated with the location",
					"items": bson.M{
						"bsonType": "string",
					},
				},
				"rating": bson.M{
					"bsonType":    "double",
					"minimum":     0,
					"maximum":     5,
					"description": "Rating of the location (0-5)",
				},
				"photos": bson.M{
					"bsonType":    "array",
					"description": "URLs to photos of the location",
					"items": bson.M{
						"bsonType": "string",
					},
				},
				"created_at": bson.M{
					"bsonType":    "date",
					"description": "Date when the location was created",
				},
				"updated_at": bson.M{
					"bsonType":    "date",
					"description": "Date when the location was last updated",
				},
			},
		},
	}

	// Create collection with validation
	opts := options.CreateCollection().SetValidator(validator)
	if err := db.CreateCollection(ctx, collectionName, opts); err != nil {
		return err
	}

	// Create indexes for better query performance
	coll := db.Collection(collectionName)

	// Index on name for text search
	nameIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: "text"}, {Key: "description", Value: "text"}},
	}

	// Index on category for filtering
	categoryIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "category", Value: 1}},
	}

	// Geospatial index for location-based queries
	geoIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}

	// Create all indexes
	_, err = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		nameIndex,
		categoryIndex,
		geoIndex,
	})

	if err != nil {
		return err
	}

	log.Printf("Collection '%s' created successfully with indexes", collectionName)
	return nil
}

// SetupUsersCollection creates the users collection with validation
func SetupUsersCollection(ctx context.Context, db *mongo.Database) error {
	// Check if collection already exists
	collectionName := "users"
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return err
	}

	for _, coll := range collections {
		if coll == collectionName {
			log.Printf("Collection '%s' already exists", collectionName)
			return nil
		}
	}

	// Define validation schema for users collection
	validator := bson.M{
		"$jsonSchema": bson.M{
			"bsonType": "object",
			"required": []string{"username", "email", "created_at"},
			"properties": bson.M{
				"username": bson.M{
					"bsonType":    "string",
					"description": "Username must be a string and is required",
				},
				"email": bson.M{
					"bsonType":    "string",
					"pattern":     "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
					"description": "Email must be a valid email address and is required",
				},
				"password_hash": bson.M{
					"bsonType":    "string",
					"description": "Hashed password for the user",
				},
				"full_name": bson.M{
					"bsonType":    "string",
					"description": "Full name of the user",
				},
				"profile_pic": bson.M{
					"bsonType":    "string",
					"description": "URL to profile picture",
				},
				"bio": bson.M{
					"bsonType":    "string",
					"description": "User biography",
				},
				"preferences": bson.M{
					"bsonType":    "object",
					"description": "User preferences",
				},
				"saved_locations": bson.M{
					"bsonType":    "array",
					"description": "Array of saved location IDs",
					"items": bson.M{
						"bsonType": "string",
					},
				},
				"created_at": bson.M{
					"bsonType":    "date",
					"description": "Date when the user was created",
				},
				"updated_at": bson.M{
					"bsonType":    "date",
					"description": "Date when the user was last updated",
				},
				"last_login": bson.M{
					"bsonType":    "date",
					"description": "Date of the user's last login",
				},
			},
		},
	}

	// Create collection with validation
	opts := options.CreateCollection().SetValidator(validator)
	if err := db.CreateCollection(ctx, collectionName, opts); err != nil {
		return err
	}

	// Create indexes for better query performance
	coll := db.Collection(collectionName)

	// Unique index on username
	usernameIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Unique index on email
	emailIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	// Create all indexes
	_, err = coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		usernameIndex,
		emailIndex,
	})

	if err != nil {
		return err
	}

	log.Printf("Collection '%s' created successfully with indexes", collectionName)
	return nil
}
