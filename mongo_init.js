// db = db.getSiblingDB('mydb');
// db.createCollection('todos', {capped: false});

const util = require('util');
db = db.getSiblingDB('mydb');
util.sleep(5000);
db.createCollection('todos', {capped: false});
