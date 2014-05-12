Mylar Distributed DB
====================

A secure distributed database, built as a Mylar/Meteor package, with end-to-end encryption and usable public key infrastructure.

Dependencies
============

- git
- mylar, either:
  - <https://github.com/jmoldow/mylar.git>@public-meteor-release/0.8.1.1 branch
    - meteor fork version: v0.8.1.1
    - Website: <https://github.com/jmoldow/mylar>
  - <git://g.csail.mit.edu/mylar>@public branch
    - meteor fork version: v0.6.6.3
    - Website: <https://css.csail.mit.edu/mylar/>
- mylar library dependencies:
  - libpbc, libpbc-dev
    - Location: /usr/local
    - Website: <https://crypto.stanford.edu/pbc/>
  - libntl, libntl-dev
  - libreadline, libreadline-dev
  - libgmp, libgmp-dev, libgmpxx, libgmpxx-dev
  - libcrypto++9, libcrypto++9-dev
  - libssl, libssl-dev
- mylar program dependencies
  - gcc, g++
  - perl
- mylar Node dependencies
  - mime
    - a Node mime-type mapping library
    - Website: <https://www.npmjs.org/package/mime>
- meteorite
  - a meteor package manager
  - Website: <https://www.npmjs.org/package/meteorite>

Installation
============

1. Manually install the mylar library dependencies, the mylar program
   dependencies, and git.

1. Use ```git clone -- https://github.com/jmoldow/MylarMail.git
   [/path/to/MylarMail]``` to checkout the MylarMail repository.

1. Use ```git clone -b <branch-name> -- <repository> [/path/to/mylar]``` to
   checkout the desired mylar repository.

1. (Optional) Run ```ln -s /path/to/mylar/meteor /usr/local/bin/mylar ``` (may
   require root privileges) to install mylar globally. If you do this, you may
   run ```mylar``` in place of ```/path/to/mylar/meteor```.

1. Run ```/path/to/mylar/meteor --get-ready``` to have meteor automatically
   install its Node dependencies.

1. Run
   ```
   cd /path/to/mylar/dev_bundle/lib/node_modules
   ../../bin/npm install mime meteorite
   ```
   to install mime and meteorite.

1. Run
   ```
   export PATH=$PATH:/path/to/mylar
   cd /path/to/MylarMail
   /path/to/mylar/dev_bundle/bin/node /path/to/mylar/dev_bundle/lib/node_modules/meteorite/bin/mrt.js --get-ready
   ```
   to have meteorite automatically install MylarMail's meteor package dependencies.

Run (Local Development Mode)
============================

1. In one shell, run
   ```
   cd /path/to/mylar/enc_modules/idp-email
   /path/to/mylar/meteor --port=3000
   ```

1. In another shell, run
   ```
   cd /path/to/MylarMail
   /path/to/mylar/meteor --port=4000
   ```

