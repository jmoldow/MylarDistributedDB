/*
 * mylar_mail_common.js - Common code for the mylar-mail Meteor package.
 */

MYLAR_CHOSEN_IDP_HOST = "";

Meteor.startup(function () {

  // On start-up, choose a random IDP server to use.
  // MYLAR_IDP_HOSTS and MYLAR_IDP_PUB must be defined in an application as
  // package-scope variables (variables with no 'var' declaration).
  // MYLAR_IDP_HOSTS  - array of strings, hosts of the IDP servers
  // MYLAR_IDP_PUB    - string, the common public key of all the IDP servers
  MYLAR_CHOSEN_IDP_HOST = Random.choice(MYLAR_IDP_HOSTS);
  idp_init(MYLAR_CHOSEN_IDP_HOST, MYLAR_IDP_PUB, false);

});
