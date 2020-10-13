const express = require('express');
const gamedig = require('gamedig');
const { check, oneOf, validationResult } = require('express-validator');

const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.send('Hello World!');
});

app.get('/server', [
  oneOf([check('ip').isIP(), check('ip').isURL()], 'Invalid ip address or url'),
  check('qport').isPort().withMessage('Invalid port number')], async (req, res) => {
  const { ip, qport } = req.query;
  try {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      res.status(400).send({ errorMessages: errors.array() });
      return;
    }
    const data = await gamedig.query({ type: 'tf2', host: ip, port: qport });
    const r = (
      { name: data.name, playerCount: data.players.length, maxPlayerCount: data.maxplayers }
    );
    res.send(r);
  } catch (err) {
    res.status(400).send(err.message);
  }
});
app.listen(port, () => {
  console.log(`Example app listening at http://localhost:${port}`);
});
