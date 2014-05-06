/*
 * main.js JavaScript files are loaded after everything else in the app.
 */

Accounts.onCreateUser(function (options, user) {
  if (options.profile)
    user.profile = options.profile;
  // The principal and accounts-idp packages assume that the username and
  // email are the same, and that only one is given. So the MylarMail client
  // only asks for an email, and the username is set to be equal to it
  // before the account is created.
  if (!user.username && (user.emails.length > 0))
    user.username = user.emails[0].address;
  return user;
});
