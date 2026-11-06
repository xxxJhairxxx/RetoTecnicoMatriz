'use strict';

const { buildApp } = require('./container');

const port = process.env.PORT || 3000;
const jwtSecret = process.env.JWT_SECRET;

if (!jwtSecret) {
  console.error('JWT_SECRET es obligatorio');
  process.exit(1);
}

buildApp({ jwtSecret }).listen(port, () => {
  console.log(`API de estadisticas escuchando en :${port}`);
});
