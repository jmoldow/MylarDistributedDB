/*
 * errors.js - Defines the Meteor.WrongConnectionError subclass of
 *             Meteor.Error, as well as the exported function
 *             CastErrorToCorrectSubclass().
 */

// 300 Multiple Choices
// <http://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html#sec10.3.1>
var WrongConnectionErrorCode = 300;

// Meteor.WrongConnectionError is thrown by the server whenever a query is made
// for a user whose data is not replicated at that server. This might happen if
// the client initially connects to the wrong server, or if the replicas move.
// It is a subclass of Meteor.Error, and contains the additional fields:
// username         {String}
// userReplicaUrls  {[String]}  Ordered list of replica servers storing the
//                              data for username. Each must be the URL of
//                              another application that can be passed
//                              to DDP.connect().
var WrongConnectionError = Meteor.makeErrorType(
  "Meteor.WrongConnectionError",

  // @param username        {String}
  // @param userReplicaUrls {[String]}  See comments above.
  function (username, userReplicaUrls) {
    var self = this;

    self.error = WrongConnectionErrorCode;

    self.reason = "Wrong Connection";

    // When the client receives an instance of Meteor.Error, it only keeps the
    // error, reason, and details fields. So we store duplicates of the
    // additional fields inside the details field, so that the full
    // WrongConnectionError can be reconstructed in the client application.
    self.details = {
      details:  "Connected to wrong server for requested username " +
                "(" + username + "). " +
                "Reconnect to one of the provided userReplicaUrls.",
      username: username,
      userReplicaUrls:  userReplicaUrls
    };

    self.message = self.reason + ' [' + self.error + ']';

    self.username = username;

    self.userReplicaUrls = userReplicaUrls;
  });

// Makes Meteor.WrongConnectionError inherit from Meteor.Error.
var tmp = function () {};
var parent = Meteor.Error;
tmp.prototype = parent.prototype;
WrongConnectionError.prototype = new tmp;
WrongConnectionError.prototype.constructor = WrongConnectionError;
Meteor.WrongConnectionError = WrongConnectionError;

// When the client receives an instance of Meteor.Error, it only keeps the
// error, reason, and details fields, and constructs a new Meteor.Error, even
// if the received object was for a subclass of Meteor.Error.
// CastErrorToCorrectSubclass() identifies if an error is an instance of
// another error class, and if so constructs and returns the error as a proper
// instance of that error class. Otherwise, the original object is returned.
// Currently supports Meteor.WrongConnectionError.
CastErrorToCorrectSubclass = function (error) {
  if ( ! (error instanceof Meteor.Error) ) {
    return error;
  }
  if ( error.error == WrongConnectionErrorCode &&
       typeof(error.details) === "object" &&
       error.details.userReplicaUrls
     ) {
    return new Meteor.WrongConnectionError(error.details.username,
                                           error.details.userReplicaUrls);
  }
  return error;
}

