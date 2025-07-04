/**
 * Import function triggers from their respective submodules:
 *
 * const {onCall} = require("firebase-functions/v2/https");
 * const {onDocumentWritten} = require("firebase-functions/v2/firestore");
 *
 * See a full list of supported triggers at https://firebase.google.com/docs/functions
 */

const functions = require("firebase-functions");
const admin = require("firebase-admin");
const {MongoClient} = require("mongodb");
const neo4j = require("neo4j-driver");

// Initialize Firebase Admin
admin.initializeApp();

// Database configuration
const MONGO_URI = "mongodb://mongodb:27017";
const DB_NAME = "authentication_service";
const NEO4J_URI = "bolt://localhost:7687";
const NEO4J_USER = "neo4j";
const NEO4J_PASSWORD = "Aswin@1996";

// For cost control, you can set the maximum number of containers that can be
// running at the same time. This helps mitigate the impact of unexpected
// traffic spikes by instead downgrading performance.
const runtimeOpts = {
  maxInstances: 10,
};

exports.syncUserToDBs = functions
    .runWith(runtimeOpts)
    .firestore
    .document("users/{userId}")
    .onCreate(async (snap, context) => {
      const data = snap.data();
      const userId = context.params.userId;

      let mongoClient;
      let neoSession;

      try {
      // Connect to MongoDB
        mongoClient = new MongoClient(MONGO_URI);
        await mongoClient.connect();

        const db = mongoClient.db(DB_NAME);
        const collection = db.collection("users");

        // Insert user data to MongoDB
        await collection.insertOne({
          uid: userId,
          createdAt: admin.firestore.FieldValue.serverTimestamp(),
          ...data,
        });

        console.log(`User ${userId} synced to MongoDB`);

        // Connect to Neo4J
        const neoDriver = neo4j.driver(
            NEO4J_URI,
            neo4j.auth.basic(NEO4J_USER, NEO4J_PASSWORD),
        );

        neoSession = neoDriver.session();

        // Create user node in Neo4J
        const createQuery = "CREATE (u:User {uid: $uid, email: $email, " +
        "phone: $phone, createdAt: datetime()})";
        await neoSession.run(createQuery, {
          uid: userId,
          email: data.Email || null,
          phone: data.PhoneNumber || null,
        });

        console.log(`User ${userId} synced to Neo4J`);
        console.log(`Successfully synced user ${userId} to both databases`);
      } catch (error) {
        console.error("Error syncing user to databases:", error);
        throw new functions.https.HttpsError(
            "internal",
            "Failed to sync user to databases",
            error.message,
        );
      } finally {
      // Clean up connections
        if (mongoClient) {
          await mongoClient.close();
        }
        if (neoSession) {
          await neoSession.close();
        }
      }
    });

// Optional: Add a simple HTTP function for testing
exports.helloWorld = functions.https.onRequest((request, response) => {
  functions.logger.info("Hello logs!", {structuredData: true});
  response.send("Hello from Firebase!");
});
