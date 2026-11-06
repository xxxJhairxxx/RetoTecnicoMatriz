'use strict';

const { createApp } = require('./app');
const { computeStatistics } = require('./domain/statistics');

function buildApp({ jwtSecret }) {
  return createApp({ jwtSecret, computeStatistics });
}

module.exports = { buildApp };
