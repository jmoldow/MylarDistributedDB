getPort = function () {
  var url = Meteor.absoluteUrl();
  var match = url.match(/:([0-9]+)/);
  if (match) {
    return match[1];
  }
}

function getPortFromSocket (socket) {
  var match = socket.match(/([0-9]+)$/);
  if (match) {
    return match[1];
  }
}

getHostFromSocket = function (socket) {
  var url = Meteor.absoluteUrl();
  url = url.replace(/:([0-9]+)/, ":"+getPortFromSocket(socket));
  return url;
}
