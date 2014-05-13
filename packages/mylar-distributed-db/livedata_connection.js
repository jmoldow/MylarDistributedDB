/*
 * livedata_connection.js - Configure clients to work in a distributed server
 *                          environment where the user might be connected with
 *                          the wrong server.
 */

// If we have a logged in user, userReplicaUrls should be set to be the array
// of replicas storing the user's data.
Meteor.userReplicaUrls = [];

// If we have connected to a Meteor server based on the logged in user,
// userCoordinatorUrl should be set to the host name of the connected server.
// At the time of assignment, it should be a member of userReplicaUrls.
Meteor.userCoordinatorUrl = "";

// Connect Meteor to a new server at the given url.
function connect (url) {

  // Clearing userCoordinatorUrl before disconnecting and reconnecting prevents
  // our disconnect handler (see below) from running during this manual
  // disconnect. This effectively locks this connect() method and the body of
  // our disconnect handler. Not doing so leads to an unbounded number of calls
  // to connect(url) (with the same url) and an eventual stack overflow error.
  // DO NOT REMOVE THIS LINE!
  Meteor.userCoordinatorUrl = "";

  // Clear the connection retryCount to 0.
  // This can also be accomplished by disconnecting before reconnecting, but
  // that causes race conditions.
  Meteor.connection._stream.currentStatus.retryCount = 0;

  // Reconnect to the new url.
  // The Meteor.connection object is mutated, not replaced, so all objects that
  // share the same reference (including the global Accounts object and all
  // defined collections) will also have the new connection.
  Meteor.connection._stream.reconnect({url: url});
  Meteor.userCoordinatorUrl = url;
}

// Save a correctly bound reference to Meteor.connection.apply() before we
// redefine it.
var apply = Meteor.bindEnvironment(Meteor.connection.apply,
                                   undefined, Meteor.connection);

// Wraps Meteor.apply() to handle WrongConnectionErrors.
//
// When an apply() is attempted, the server may reply with
// WrongConnectionError, indicating that it is not a replica for the current
// user. The client should cache the list of replicas, connect to the first
// one, and re-send the apply().
Meteor.connection.apply = function (name, args, options, callback, meta) {
  // Deal with the possibility of receiving less than five arguments.  Copied
  // from the implementation of apply() in
  // packages/livedata/livedata_connection.js in the mylar source.
  if (!callback && typeof options === 'function') {
    callback = options;
    options = {};
  }
  options = options || {};
  if (callback) {
    callback = Meteor.bindEnvironment(
      callback,
      "delivering result of invoking '" + name + "'"
    );
  }

  // A wrapped callback to handle WrongConnectionErrors.
  var connectionUpdateCallback = function (error, result) {
    error = CastErrorToCorrectSubclass(error);
    if (error && (error instanceof Meteor.WrongConnectionError)) {
      Meteor.userReplicaUrls = error.userReplicaUrls;
      connect(error.userReplicaUrls[0]);

      // Recurse after changing connection.
      Meteor.connection.apply(name, args, options, callback, meta);
    }
    else if (callback) {
      callback(error, result);
    }
  }

  // Call original apply(), but with wrapped callback.
  apply(name, args, options, connectionUpdateCallback, meta);
}

// If the connected server becomes unreachable and we know what replicas are
// storing the current user's data, establish a connection with one of the
// other replicas.
Meteor.connection._stream.on('disconnect', function () {
  // Only switch replicas if we are not currently performing a
  // disconnect/reconnect (see the comments in the definition of connect()), if
  // we have a list of replicas to choose from, and if we have tried connecting
  // to the current server at least twice.
  if ( Meteor.userCoordinatorUrl &&
       Meteor.userReplicaUrls.length > 0 &&
       Meteor.status().retryCount > 1 ) {
    // Choose the next replica in the list, circling around if we've reached
    // the end, and starting at the beginning if the current server is not in
    // the list (i.e., if indexOf() returns -1).
    // Each failed connection try actually triggers a disconnect event. So
    // there is no need to do any iteration in this handler (in fact, it
    // wouldn't even work). This handler will be called repeatedly until a
    // suitable connection has been made.
    var i = 1 + _.indexOf(Meteor.userReplicaUrls, Meteor.userCoordinatorUrl);
    i = i % Meteor.userReplicaUrls.length;
    connect(Meteor.userReplicaUrls[i]);
  }
});

// When the client is loaded, Meteor.userReplicaUrls is computed based on
// the current user. And since Meteor.userId() is reactive, it is recomputed
// whenever the current user changes.
Deps.autorun(function () {
  Meteor.apply('getUserReplicaUrls',
               [Meteor.userId()],
               {wait: true},
               function (error, result) {
    if (result && !error) {
      Meteor.userReplicaUrls = result;
      if (result.length > 0) {
        // If we are not currently connected to a coordinator, make a new
        // connection immediately.
        if (!Meteor.userCoordinatorUrl) {
          connect(result[0]);
        }
      }
      else {
        Meteor.userCoordinatorUrl = "";
      }
    }
  });
});

var _getUserReplicaUrls = function (userId, callback) {
  Meteor.apply('getUserReplicaUrls', [userId], callback);
}

_getUserReplicaUrls = Meteor._wrapAsync(_getUserReplicaUrls);

getUserReplicaUrls = function (userId) {
  return _getUserReplicaUrls(userId);
}
