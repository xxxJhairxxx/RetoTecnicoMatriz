'use strict';

const { ValidationError } = require('../domain/statistics');

function makeStatisticsController(computeStatistics) {
  return (req, res) => {
    const { q, r } = req.body ?? {};
    try {
      res.json(computeStatistics({ q, r }));
    } catch (err) {
      if (err instanceof ValidationError) return res.status(422).json({ error: err.message });
      throw err;
    }
  };
}

module.exports = { makeStatisticsController };
