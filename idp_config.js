/*
 * idp_config.js - Application configuration for IDP servers.
 */

MYLAR_ACTIVE_ATTACKER = true;
// use IDP only if active attacker
Accounts.config({sendVerificationEmail:active_attacker(),
                 loginExpirationInDays:1});

MYLAR_IDP_PUB = '8a7fe03431b5fc2db3923a2ab6d1a5ddf35cd64aea35e743' +
                'ded7655f0dc7e085858eeec06e1c7da58c509d57da56dbe6';
MYLAR_IDP_HOSTS = ["http://localhost:3000"]

// Choose a random IDP server to use.
// MYLAR_IDP_HOSTS and MYLAR_IDP_PUB must be defined in an application as
// package-scope variables (variables with no 'var' declaration).
// MYLAR_IDP_HOSTS  - array of strings, hosts of the IDP servers
// MYLAR_IDP_PUB    - string, the common public key of all the IDP servers
MYLAR_CHOSEN_IDP_HOST = Random.choice(MYLAR_IDP_HOSTS);
idp_init(MYLAR_CHOSEN_IDP_HOST, MYLAR_IDP_PUB, false);
