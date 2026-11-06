'use strict';

const ZERO_TOLERANCE = 1e-9;

class ValidationError extends Error {}

function assertMatrix(value, name) {
  if (!Array.isArray(value) || value.length === 0 || !Array.isArray(value[0]) || value[0].length === 0) {
    throw new ValidationError(`"${name}" debe ser un array de arrays no vacio`);
  }
  const columns = value[0].length;
  for (const row of value) {
    if (!Array.isArray(row) || row.length !== columns) {
      throw new ValidationError(`"${name}" debe ser rectangular: todas las filas con ${columns} columnas`);
    }
    for (const cell of row) {
      if (typeof cell !== 'number' || !Number.isFinite(cell)) {
        throw new ValidationError(`"${name}" solo admite numeros finitos, se recibio ${JSON.stringify(cell)}`);
      }
    }
  }
  return value;
}

function isDiagonal(matrix) {
  return matrix.every((row, i) => row.every((cell, j) => i === j || Math.abs(cell) <= ZERO_TOLERANCE));
}

function computeStatistics(matrices) {
  const entries = Object.entries(matrices);
  let max = -Infinity;
  let min = Infinity;
  let sum = 0;
  let count = 0;
  const diagonal = {};

  for (const [name, matrix] of entries) {
    assertMatrix(matrix, name);
    diagonal[name] = isDiagonal(matrix);
    for (const row of matrix) {
      for (const cell of row) {
        if (cell > max) max = cell;
        if (cell < min) min = cell;
        sum += cell;
        count += 1;
      }
    }
  }

  if (count === 0) throw new ValidationError('no se recibio ninguna matriz');

  return {
    max,
    min,
    sum,
    average: sum / count,
    count,
    diagonal,
    anyDiagonal: Object.values(diagonal).some(Boolean),
  };
}

module.exports = { computeStatistics, isDiagonal, assertMatrix, ValidationError, ZERO_TOLERANCE };
