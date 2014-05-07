/*
 * router.js - iron-router routing table.
 */

Router.configure({
});

// 1. waiting filter
Router.onBeforeAction(function() {
  if (! this.ready()) { }
});

Router.map(function() {
    this.route('home', {path: '/'})
    this.route('about');
});
