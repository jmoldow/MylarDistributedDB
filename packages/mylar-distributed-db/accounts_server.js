/*
 * accounts_server.js - The distributed database chooses replicas for a user's
 *                      data by taking a hash of that user's username. There are
 *                      situation where a client may be connected to a server
 *                      which is not one of the replicas for the requested
 *                      username (during account creation, account login,
 *                      password change, session resume, and any other scenario
 *                      involving a unique user on the client). When a client
 *                      executes a query on the connected server which is for a
 *                      user not replicated on that server, the server needs to
 *                      be able to identify this and signal the user. The
 *                      following code enables this to happen.
 */

Accounts.onCreateUser(function (options, user) {
  if (options.profile)
    user.profile = options.profile;
  // The principal and accounts-idp packages assume that the username and email
  // are the same, and that only one is given by the client.  They also need to
  // be the same in the distributed environment, so that findOne() queries by
  // email address can be directed at a proper replica for that user.
  // So clients should only ask for an email for account creation.
  // The username is set to be equal to it before the account is created.
  if (!user.username && (user.emails.length > 0))
    user.username = user.emails[0].address;

  // The distributed database chooses replicas for a user's data by taking a
  // hash of that user's username. When a client executes a query on the
  // connected server which is for a user not replicated on that server, the
  // server should throw an WrongConnectionError and redirect the client to the
  // correct replica servers. However, this requires that the query has enough
  // information in it to always idenfiy the username, even when that user is
  // not stored on that server. When the query is by _id only and the _id is a
  // random string, this is impossible to do. By making the _id not random, but
  // instead making _id === username === emails[0].address, we make it always
  // possible to know the username, and therefore the username hash, associated
  // with a user query, regardless of whether the query is by _id, username, or
  // email.  The only restriction on _id is that it must be unique. Since
  // username is already guaranteed to be unique, this is not a problem.
  user._id = user.username;

  return user;
});

Meteor.methods({
  // A server method that computes and returns the ordered replica list for the
  // given userId (=== username).
  // If userId is not given, it defaults to the userId of the connected user.
  // If this is undefined, then the empty list is returned.
  getUserReplicaUrls: function (userId) {
    if (!userId) {
      userId = this.userId;
    }
    if (userId) {
      coordinators = GetCoordinatorList(userId);
      return coordinators
    }
    return [];
  }});

// On the server, wrap all of the user collection functions.
//
// If the operation is for a single user, first check if this server is a
// replica for that user. This is possible to do without any querying, because
// - single user queries will almost always be by id, username, or email
// - a user's _id === username === emails[0].address
// - the replicas for a user are determined by a hash of the username
// If this is a replica for the user, perform the query. Otherwise, throw
// WrongConnectionError.
//
// Wrapping these methods means, in particular, that clients will be directed
// towards the correct replicas on account creation and login, which use
// insert() and findOne() respectively. There is no need to redefine every
// possible account creation / login method, and no need to write
// application-specific code to handle account operations.
_.each(["insert", "update", "upsert",
        "remove", "find", "findOne"], function (name) {
  var fn = Meteor.users[name];
  fn = Meteor.bindEnvironment(fn, Meteor._debug, Meteor.users);
  Meteor.users[name] = function (/*arguments*/) {
    selector = arguments.length ? arguments[0] : {};
    var username;
    // For users, _id === username === emails[0].address
    if (typeof selector === "string") {
      username = selector;
    }
    else if (typeof selector === "object") {
      if (selector._id) {
        username = selector._id;
      }
      else if (selector.id) {
        username = selector.id;
      }
      else if (selector.username) {
        username = selector.username;
      }
      else if (selector["emails.address"]) {
        username = selector["emails.address"];
      }
      else if (selector.emails) {
        try {
          username = selector.emails[0].address;
        } catch (e) { }
      }
    }
    if (username) {
      // TODO call GetCoordinatorList(username)
      // coordinators = ["http://localhost:6000", "http://localhost:5000", "http://localhost:4000", "http://localhost:3000"]; // for testing purposes
      // throw new Meteor.WrongConnectionError(username, coordinators);
    }
    return fn.apply(this, arguments);
  }
});
