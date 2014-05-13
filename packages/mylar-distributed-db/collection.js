/*
 * collection.js - Utilities for wrapping collection operations so that
 *                 operations will not directly be applied to the server's
 *                 underlying MongoDB database, but instead will be sent
 *                 through the distributed database.
 */

Meteor.collections = {};

wrap_insert = function (collection, getUserId, onflict_resolution) {
  Meteor.collections[collection._name] = collection;
  if (Meteor.isServer) {
    if (!conflict_resolution || !(conflict_resolution instanceof Function)) {
      // Default conflict resolution is to take the object with the latest
      // timestamp.
      conflict_resolution = function (doc1, doc2) {
        return (doc1._ts >= doc2._ts ? doc1 : doc2);
      }
    }
    collection.localFindOne = collection.findOne;
    collection._collection.localUpdate = collection._collection.update;
    var localInsert = collection._collection.insert;
    localInsert = Meteor._wrapAsync(Meteor.bindEnvironment(localInsert, Meteor._debug, collection._collection));
    collection.localPut = Meteor._wrapAsync(function (doc, callback) {
      console.log(collection._name + ".localPut defined");
      // If an object currently exists, find it, perform conflict
      // resolution, and update the document.
      // Otherwise, insert the new document.
      var docCurrent = collection.localFindOne({_id: doc._id});
      if (docCurrent) {
        doc = conflict_resolution(docCurrent, doc);
        return collection._collection.localUpdate(doc._id, doc, function (error) { return callback(error, doc._id); });
      }
      else {
        return localInsert(doc, callback);
      }
    });
    console.log(collection._name + ".localPut defined");
    collection._collection.localRemove = collection._collection.remove;
    collection._collection.remove = Meteor._wrapAsync(function (id, callback) {
      console.log(collection._name + ".remove redefined");
      // We need to keep track of deleted objects. So a remove is actually
      // just overwriting the object with an empty object.
      return collection._collection.insert({_id: id}, function (error) { return callback(error, 1); });
    });
    console.log(collection._name + ".remove redefined");
    collection._collection.insert = Meteor._wrapAsync(function (doc, callback) {
      console.log(collection._name + ".insert redefined");
      // The server method that gets called by the client. Instead of
      // inserting directly into MongoDB, call out to the distributed
      // database. But first, give the object an _id (so that all replicas
      // will get the same _id for this object) and a timestamp (for
      // conflict resolution).
      doc = _.clone(doc);
      if (!doc._id) {
        doc._id = Random.id();
      }
      doc._ts = new Date();
      reply = CoordinatorPut(collection._name, getUserId(doc), doc._id, doc);
      console.log("about to finish collection.insert with " + reply);
      if ("OK" === reply) {
        return collection.localPut(doc, callback);
      }
    });
    console.log(collection._name + ".insert redefined");
  }
  else if (Meteor.isClient) {
    _.each(["update", "upsert"], function (name) {
      collection[name] = function () {
        throw new Meteor.Error(405, "Method Not Allowed",
                               collection._name+"."+name+"() not allowed. " +
                               "Instead call "+collection._name+".insert().");
      }
    });
  }
};

wrap_insert(
  Meteor.users,
  function (doc) { return doc._id; },
  function (doc1, doc2) {
    // use the latest version, but always keep the certificate
    var doc = _.clone(doc1._ts >= doc2._ts ? doc1 : doc2);
    var other = (doc1._ts >= doc2._ts ? doc2 : doc1);
    if (!doc._pubkey_cert) {
      doc._pubkey_cert = other._pubkey_cert;
      doc.emails[0].verified = other.emails[0].verified;
    }
});

wrap_insert(Principals, function (doc) { return doc.name; });

// not WrappedKeys, Certs, PrincType PrincType doesn't need duplication (the
// objects are static for an application), and the others aren't supported.
