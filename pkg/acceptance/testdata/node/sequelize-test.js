// Copyright 2019 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

const assert = require('assert');
const fs = require('fs');

const Sequelize = require('sequelize-cockroachdb');
const Op = Sequelize.Op;

const config = {
  dialect: 'postgres',
  host: process.env.PGHOST || 'localhost',
  port: process.env.PGPORT || 26257,
  logging: false,
};

if (process.env.PGSSLCERT && process.env.PGSSLKEY) {
  config.ssl = true;
  config.dialectOptions = {
    ssl: {
      cert: fs.readFileSync(process.env.PGSSLCERT),
      key: fs.readFileSync(process.env.PGSSLKEY)
    }
  };
}

const sequelize = new Sequelize('node_test', 'root', '', config);

describe('sequelize', () => {
  after(() => {
    sequelize.close();
  });

  it('can create a model with a json field', () => {
    var Cat = sequelize.define('cat', {
      id: {type: Sequelize.INTEGER, primaryKey: true},
      data: {type: Sequelize.JSONB},
    });

    return Cat.sync({force: true})
      .then(() => {
        return Cat.bulkCreate([
          {id: 1, data: {name: 'smudge'}},
          {id: 2, data: {name: 'sissel'}},
        ]);
      })
      .then(() => {
        return Cat.findAll();
      })
      .then(result => {
        assert.deepEqual(result[0].dataValues.id, 1);
        assert.deepEqual(result[0].dataValues.data, {name: 'smudge'});
        assert.deepEqual(result[1].dataValues.id, 2);
        assert.deepEqual(result[1].dataValues.data, {name: 'sissel'});
      });
  });

  it('can create a model with an inverted index', () => {
    var Android = sequelize.define(
      'androids',
      {
        id: {type: Sequelize.INTEGER, primaryKey: true},
        data: {type: Sequelize.JSONB},
      },
      {
        // Not sure of a good but not fragile way to verify that this index was
        // actually created.
        indexes: [
          {
            fields: ['data'],
            using: 'gin',
          },
        ],
      }
    );

    return Android.sync({force: true})
      .then(() => {
        return Android.bulkCreate([
          {id: 1, data: {name: '2B'}},
          {id: 2, data: {name: '9S'}},
        ]);
      })
      .then(() => {
        return Android.findAll({
          where: {
            data: {
              [Op.contains]: {name: '2B'},
            },
          },
        });
      })
      .then(result => {
        assert.deepEqual(result[0].dataValues.id, 1);
        assert.deepEqual(result[0].dataValues.data, {name: '2B'});
      });
  });
});
