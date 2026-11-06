'use strict';

function errorHandler(err, _req, res, _next) {
  const status = err.status && err.status < 500 ? err.status : 500;
  res.status(status).json({ error: err.message || 'error interno' });
}

module.exports = { errorHandler };
