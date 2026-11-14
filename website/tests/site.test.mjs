import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {selection} from '../selection.js';
import {languages,platforms,guide} from '../content.mjs';
test('selection validates parameters independently',()=>{assert.deepEqual(selection('?lang=rust&os=windows'),{lang:'rust',os:'windows'});assert.deepEqual(selection('?lang=bad&os=macos'),{lang:'cpp',os:'macos'});assert.deepEqual(selection('?lang=ocaml&os=bad'),{lang:'ocaml',os:'linux'})});
test('every variant contains seven complete steps and minimal configuration',()=>{for(const l of Object.keys(languages))for(const o of Object.keys(platforms)){const steps=guide(l,o);assert.equal(steps.length,7);assert.match(steps[3].configuration,/editor: ""/);assert.match(steps[3].configuration,/include: \{\}/);for(const s of steps){assert.ok(s.title&&s.intro&&s.expected&&s.help)}}});
test('static output retains all variants without initial video source',()=>{const home=readFileSync(new URL('../dist/index.html',import.meta.url),'utf8');const start=readFileSync(new URL('../dist/get-started/index.html',import.meta.url),'utf8');assert.equal((start.match(/class="variant"/g)||[]).length,12);assert.equal((start.match(/class="step"/g)||[]).length,84);assert.doesNotMatch(home,/<video|<source|\.gif/);assert.match(home,/data-theme="dark"/);assert.match(home,/captions.vtt/)});
