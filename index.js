const express = require('express');
const gamedig = require('gamedig');
const { check, oneOf, validationResult } = require('express-validator');
const redisClient = require('./redis-client');

const app = express();
const PORT = process.env.PORT || 3000;

app.get('/', (req, res) => {
  res.send('Hello World!');
});

app.get('/server', [
  oneOf([check('ip').isIP(), check('ip').isURL()], 'Invalid ip address or url'),
  check('port').isPort().withMessage('Invalid port number')], async (req, res) => {
  const { ip, port } = req.query;
  try {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
      res.status(400).send({ errorMessages: errors.array() });
      return;
    }
    const key = `server:${ip}/${port}`;
    const expire = 30;
    const data = await redisClient.getAsync(key);
    if (!data) {
      const data = await gamedig.query({ type: 'tf2', host: ip, port: port });
      const r = (
        { name: data.name, playerCount: data.players.length, maxPlayerCount: data.maxplayers }
      );
      await redisClient.setAsync(key, JSON.stringify(r), 'EX', expire);
      res.send(r);
    }
    else {
      const r = JSON.parse(data);
      res.send(r);
    }    
    
  } catch (err) {
    res.status(400).send(err.message);
  }
});

app.listen(PORT, () => {
  console.log(`Example app listening at http://localhost:${PORT}`);
});
