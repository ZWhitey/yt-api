const express = require('express');
const gamedig = require('gamedig');
const { check, oneOf, validationResult } = require('express-validator');
const redisClient = require('./redis-client');
const mysql = require('mysql2');

const app = express();
const PORT = process.env.PORT || 3000;

const con = mysql.createConnection(
  {
    host: process.env.DBHOST,
    port: process.env.DBPORT,
    user: process.env.DBUSER,
    password: process.env.DBPASS,
    database: process.env.DBPLAYER,
  },
);

app.get('/', (req, res) => {
  res.send('Hello World!');
});

app.get('/player', [
  check('steamid').matches(/^STEAM_(0|1):(0|1):\d+$/, 'g').withMessage('Invalid SteamID'),
  check('limit').isInt({ gt: 0 }).withMessage('limit should greater than 0')], async (req, res) => {
  const errors = validationResult(req);
  if (!errors.isEmpty()) {
    res.status(400).send({ errorMessages: errors.array() });
    return;
  }
  const { steamid, limit } = req.query;
  const sql = `SELECT * FROM player_analytics WHERE auth='${steamid}' limit ${limit}`;
  try {
    const [rows] = await con.promise().query(sql);
    res.send(rows);
  } catch (err) {
    res.status(400).send(err.message);
  }
});

app.get('/summary', async (req, res) => {
  const sql = 'SELECT count(distinct(auth)) as unique_player,count(*) as total_record,count(distinct(country)) as unique_country FROM player_analytics';
  try {
    const [rows] = await con.promise().query(sql);
    res.send(rows);
  } catch (err) {
    res.status(400).send(err.message);
  }

})

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
    let data = await redisClient.getAsync(key);
    if (!data) {
      data = await gamedig.query({ type: 'tf2', host: ip, port });
      const r = (
        { name: data.name, playerCount: data.players.length, maxPlayerCount: data.maxplayers }
      );
      await redisClient.setAsync(key, JSON.stringify(r), 'EX', expire);
      res.send(r);
    } else {
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
