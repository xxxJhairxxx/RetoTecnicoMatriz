'use strict';

const express = require('express');
const { statisticsRoutes } = require('./routes/statistics.routes');
const { errorHandler } = require('./middlewares/error-handler');

function createApp({ jwtSecret, computeStatistics }) {
  const app = express();
  app.use(express.json({ limit: '5mb' }));

  app.get('/health', (_req, res) => res.json({ status: 'ok' }));
  app.use('/api/v1/statistics', statisticsRoutes({ jwtSecret, computeStatistics }));
  app.use(errorHandler);

  return app;
}

module.exports = { createApp };
