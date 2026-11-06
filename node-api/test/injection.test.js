'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const jwt = require('jsonwebtoken');
const { createApp } = require('../src/app');
const { ValidationError } = require('../src/domain/statistics');

const SECRET = 'test-secret';
const token = jwt.sign({ sub: 'demo' }, SECRET, { algorithm: 'HS256', expiresIn: '1h' });

async function start(computeStatistics) {
  const server = await new Promise((resolve) => {
    const s = createApp({ jwtSecret: SECRET, computeStatistics }).listen(0, () => resolve(s));
  });
  return {
    url: `http://127.0.0.1:${server.address().port}`,
    close: () => new Promise((resolve) => server.close(resolve)),
  };
}

function postStatistics(url, body) {
  return fetch(`${url}/api/v1/statistics`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
}

test('the HTTP layer uses the injected use case', async (t) => {
  let received = null;
  const server = await start((matrices) => {
    received = matrices;
    return { max: 9, min: 9, sum: 9, average: 9, count: 1, diagonal: { q: true, r: true }, anyDiagonal: true };
  });
  t.after(server.close);

  const res = await postStatistics(server.url, { q: [[1]], r: [[2]] });
  assert.equal(res.status, 200);
  assert.deepEqual(await res.json(), {
    max: 9, min: 9, sum: 9, average: 9, count: 1, diagonal: { q: true, r: true }, anyDiagonal: true,
  });
  assert.deepEqual(received, { q: [[1]], r: [[2]] });
});

test('translates a domain ValidationError into 422 without knowing its cause', async (t) => {
  const server = await start(() => {
    throw new ValidationError('matriz de prueba invalida');
  });
  t.after(server.close);

  const res = await postStatistics(server.url, { q: [[1]], r: [[2]] });
  assert.equal(res.status, 422);
  assert.deepEqual(await res.json(), { error: 'matriz de prueba invalida' });
});

test('an unexpected domain error becomes a 500', async (t) => {
  const server = await start(() => {
    throw new Error('unexpected failure');
  });
  t.after(server.close);

  const res = await postStatistics(server.url, { q: [[1]], r: [[2]] });
  assert.equal(res.status, 500);
});
