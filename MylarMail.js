MYLAR_ACTIVE_ATTACKER = true;
// use IDP only if active attacker
Accounts.config({sendVerificationEmail:active_attacker(),
                 loginExpirationInDays:1});

if (Meteor.isClient) {
  Template.hello.greeting = function () {
    return "Welcome to MylarMail.";
  };

  Template.hello.events({
    'click input' : function () {
      // template data, if any, is available in 'this'
      if (typeof console !== 'undefined')
        console.log("You pressed the button");
    }
  });
}

if (Meteor.isServer) {
  Meteor.startup(function () {
    // code to run on server at startup
  });
}
