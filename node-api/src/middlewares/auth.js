'use strict';

const jwt = require('jsonwebtoken');

function requireJWT(secret) {
  return (req, res, next) => {
    const header = req.get('authorization') || '';
    if (!header.startsWith('Bearer ')) {
      return res.status(401).json({ error: 'falta el header Authorization: Bearer <token>' });
    }
    try {
      jwt.verify(header.slice('Bearer '.length), secret, { algorithms: ['HS256'] });
      return next();
    } catch {
      return res.status(401).json({ error: 'token invalido o expirado' });
    }
  };
}

module.exports = { requireJWT };
