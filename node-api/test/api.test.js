'use strict';

const test = require('node:test');
const assert = require('node:assert/strict');
const jwt = require('jsonwebtoken');
const { buildApp } = require('../src/container');

const SECRET = 'test-secret';
const token = jwt.sign({ sub: 'demo' }, SECRET, { algorithm: 'HS256', expiresIn: '1h' });

async function start() {
  const server = await new Promise((resolve) => {
    const s = buildApp({ jwtSecret: SECRET }).listen(0, () => resolve(s));
  });
  return {
    url: `http://127.0.0.1:${server.address().port}`,
    close: () => new Promise((resolve) => server.close(resolve)),
  };
}

async function postStatistics(url, body, bearer = token) {
  const headers = { 'Content-Type': 'application/json' };
  if (bearer) headers.Authorization = `Bearer ${bearer}`;
  const res = await fetch(`${url}/api/v1/statistics`, { method: 'POST', headers, body });
  return { status: res.status, json: await res.json() };
}

test('POST /api/v1/statistics returns the statistics', async (t) => {
  const server = await start();
  t.after(server.close);

  const { status, json } = await postStatistics(server.url, JSON.stringify({ q: [[1, 0], [0, 1]], r: [[2, 3], [0, 4]] }));
  assert.equal(status, 200);
  assert.equal(json.sum, 11);
  assert.equal(json.max, 4);
  assert.equal(json.min, 0);
  assert.equal(json.count, 8);
  assert.deepEqual(json.diagonal, { q: true, r: false });
  assert.equal(json.anyDiagonal, true);
});

test('GET /health responds ok without authentication', async (t) => {
  const server = await start();
  t.after(server.close);

  const res = await fetch(`${server.url}/health`);
  assert.equal(res.status, 200);
  assert.deepEqual(await res.json(), { status: 'ok' });
});

test('rejects requests without a token', async (t) => {
  const server = await start();
  t.after(server.close);

  const { status } = await postStatistics(server.url, JSON.stringify({ q: [[1]], r: [[1]] }), null);
  assert.equal(status, 401);
});

test('rejects a token signed with another secret', async (t) => {
  const server = await start();
  t.after(server.close);

  const foreignToken = jwt.sign({ sub: 'intruder' }, 'another-secret', { algorithm: 'HS256' });
  const { status } = await postStatistics(server.url, JSON.stringify({ q: [[1]], r: [[1]] }), foreignToken);
  assert.equal(status, 401);
});

test('rejects an expired token', async (t) => {
  const server = await start();
  t.after(server.close);

  const expiredToken = jwt.sign({ sub: 'demo' }, SECRET, { algorithm: 'HS256', expiresIn: '-1s' });
  const { status } = await postStatistics(server.url, JSON.stringify({ q: [[1]], r: [[1]] }), expiredToken);
  assert.equal(status, 401);
});

test('responds 422 for an invalid matrix', async (t) => {
  const server = await start();
  t.after(server.close);

  const { status, json } = await postStatistics(server.url, JSON.stringify({ q: [[1, 2], [3]], r: [[1]] }));
  assert.equal(status, 422);
  assert.match(json.error, /rectangular/);
});

test('responds 400 for malformed JSON', async (t) => {
  const server = await start();
  t.after(server.close);

  const { status } = await postStatistics(server.url, '{"q": [[1,');
  assert.equal(status, 400);
});
