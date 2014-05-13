/*
 * collection.js - Utilities for wrapping collection operations so that
 *                 operations will not directly be applied to the server's
 *                 underlying MongoDB database, but instead will be sent
 *                 through the distributed database.
 */

wrap_insert = function (collection, getUserId, onflict_resolution) {
  if (Meteor.isServer) {
    if (!conflict_resolution || !(conflict_resolution instanceof Function)) {
      // Default conflict resolution is to take the object with the latest
      // timestamp.
      conflict_resolution = function (doc1, doc2) {
        return (doc1._ts >= doc2._ts ? doc1 : doc2);
      }
    }
    var localInsert = collection.insert;
    localInsert = Meteor.bindEnvironment(localInsert, Meteor._debug, collection);
    collection.localPut = function (doc/*, callback*/) {
      // If an object currently exists, find it, perform conflict
      // resolution, and update the document.
      // Otherwise, insert the new document.
      var docCurrent = collection.findOne({_id: doc._id});
      if (docCurrent) {
        doc = conflict_resolution(docCurrent, doc);
        return collection.update(doc._id, doc);
      }
      else {
        return localInsert(doc);
      }
    }
    collection.remove = function (id/*, callback*/) {
      // We need to keep track of deleted objects. So a remove is actually
      // just overwriting the object with an empty object.
      return collection.insert({_id: _id});
    }
    collection.insert = function (doc/*, callback*/) {
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
      CoordinatorPut(collection._name, getUserId(doc), doc._id, doc);
    }
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
