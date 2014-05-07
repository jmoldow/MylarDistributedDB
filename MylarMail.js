MYLAR_ACTIVE_ATTACKER = true;
// use IDP only if active attacker
Accounts.config({sendVerificationEmail:active_attacker(),
                 loginExpirationInDays:1});

if (Meteor.isClient) {
}

if (Meteor.isServer) {
  Meteor.startup(function () {
    // code to run on server at startup
  });
}
