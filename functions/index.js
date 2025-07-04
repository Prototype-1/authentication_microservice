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
const { MongoClient } = require("mongodb");
const neo4j = require("neo4j-driver");

admin.initializeApp();

// Database config
const MONGO_URI = "mongodb://localhost:27017";
const DB_NAME = "authentication_service";
//const NEO4J_URI = "neo4j://127.0.0.1:7687";
const NEO4J_USER = "neo4j";
const NEO4J_PASSWORD = "Aswin@1996";

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
      mongoClient = new MongoClient(MONGO_URI);
      await mongoClient.connect();
      const db = mongoClient.db(DB_NAME);
      const collection = db.collection("users");

      await collection.insertOne({
        uid: userId,
        createdAt: new Date(),
        ...data,
      });

      console.log(`User ${userId} synced to MongoDB`);

        const neoDriver = neo4j.driver(
  "bolt://127.0.0.1:7687",
  neo4j.auth.basic(NEO4J_USER, NEO4J_PASSWORD),
  { encrypted: "ENCRYPTION_OFF" }
);

      neoSession = neoDriver.session();

      const createQuery = `
        CREATE (u:User {
          uid: $uid,
          email: $email,
          phone: $phone,
          createdAt: datetime()
        })
      `;

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
        error.message
      );
    } finally {
      if (mongoClient) await mongoClient.close();
      if (neoSession) await neoSession.close();
    }
  });

exports.helloWorld = functions.https.onRequest((req, res) => {
  functions.logger.info("Hello logs!", { structuredData: true });
  res.send("Hello from Firebase!");
});
