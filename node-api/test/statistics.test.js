'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const { computeStatistics, isDiagonal, ValidationError } = require('../src/domain/statistics');

test('aggregates max, min, sum and average across both matrices', () => {
  const stats = computeStatistics({
    q: [[1, 2], [3, 4]],
    r: [[5, 6], [7, 8]],
  });
  assert.equal(stats.max, 8);
  assert.equal(stats.min, 1);
  assert.equal(stats.sum, 36);
  assert.equal(stats.count, 8);
  assert.equal(stats.average, 4.5);
});

test('handles negative and decimal values', () => {
  const stats = computeStatistics({ q: [[-2.5, 0.5]], r: [[1.5]] });
  assert.equal(stats.max, 1.5);
  assert.equal(stats.min, -2.5);
  assert.equal(stats.sum, -0.5);
  assert.equal(stats.average, -0.5 / 3);
});

test('detects a diagonal matrix and sets anyDiagonal', () => {
  const stats = computeStatistics({ q: [[1, 2], [3, 4]], r: [[5, 0], [0, 8]] });
  assert.deepEqual(stats.diagonal, { q: false, r: true });
  assert.equal(stats.anyDiagonal, true);
});

test('anyDiagonal is false when no matrix is diagonal', () => {
  const stats = computeStatistics({ q: [[1, 2], [3, 4]], r: [[5, 6], [0, 8]] });
  assert.equal(stats.anyDiagonal, false);
});

test('isDiagonal tolerates floating point noise', () => {
  assert.equal(isDiagonal([[1, 1e-15], [-1e-15, 2]]), true);
  assert.equal(isDiagonal([[1, 1e-6], [0, 2]]), false);
});

test('isDiagonal accepts non-square matrices', () => {
  assert.equal(isDiagonal([[1, 0], [0, 2], [0, 0]]), true);
  assert.equal(isDiagonal([[1, 0], [0, 2], [0, 3]]), false);
});

test('the null matrix is diagonal', () => {
  assert.equal(isDiagonal([[0, 0], [0, 0]]), true);
});

test('rejects inputs that are not valid matrices', () => {
  const invalidInputs = [
    { q: [], r: [[1]] },
    { q: [[]], r: [[1]] },
    { q: [[1, 2], [3]], r: [[1]] },
    { q: [[1, 'a']], r: [[1]] },
    { q: [[1, null]], r: [[1]] },
    { q: [[1, Infinity]], r: [[1]] },
    { q: [[1, NaN]], r: [[1]] },
    { q: undefined, r: [[1]] },
    { q: [[1]], r: 'not a matrix' },
  ];
  for (const input of invalidInputs) {
    assert.throws(() => computeStatistics(input), ValidationError, `should reject ${JSON.stringify(input)}`);
  }
});
