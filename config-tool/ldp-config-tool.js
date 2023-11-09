#!/usr/bin/env node

import fs from 'fs';
import Folio from '@indexdata/foliojs';

const op = process.argv[2];
if (process.argv.length !== 3 || (op !== 'get' && op !== 'set')) {
  console.error('Usage:', process.argv[1], '<get|set>');
  process.exit(1);
}

const [_, session] = await Folio.defaultSetup();
if (op === 'get') {
  const body = await session.folioFetch('/ldp/config');
  console.log(JSON.stringify(body, null, 2));
} else {
  var data = fs.readFileSync(process.stdin.fd, 'utf-8');
  const json = JSON.parse(data)
  for (let i = 0; i < json.length; i++) {
    const rec = json[i]
    await session.folioFetch(`/ldp/config/${rec.key}`, {
      method: 'PUT',
      json: rec,
    });
    console.log('record', i, '--', rec);
  }
}

session.close();
