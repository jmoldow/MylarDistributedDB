/*
 * main.js JavaScript files are loaded after everything else in the app.
 */

// The principal and accounts-idp packages assume that the username and
// email are the same, and that only one is given. So the MylarMail client
// only asks for an email, and the username is set to be equal to it before
// the account is created.
Accounts.ui.config({passwordSignupFields: 'EMAIL_ONLY'});
