private
=======

Meteor gathers any files under the private subdirectory and makes the
contents of these files available to server code via the Assets API. The
private subdirectory is the place for any files that should be accessible to
server code but not served to the client, like private data files.

Meteor excludes anything under this directory when gathering files for the
client and server.
