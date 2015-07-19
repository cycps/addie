var hooks = require('hooks');

hooks.beforeAll(function (transactions, done) {
  hooks.log('before all');
  process.env.NODE_TLS_REJECT_UNAUTHORIZED = "0";
  done();
});
