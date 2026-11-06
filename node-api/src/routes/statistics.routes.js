'use strict';

const { Router } = require('express');
const { requireJWT } = require('../middlewares/auth');
const { makeStatisticsController } = require('../controllers/statistics.controller');

function statisticsRoutes({ jwtSecret, computeStatistics }) {
  const router = Router();
  router.post('/', requireJWT(jwtSecret), makeStatisticsController(computeStatistics));
  return router;
}

module.exports = { statisticsRoutes };
