import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync,existsSync} from 'node:fs';
import {fileURLToPath} from 'node:url';
import {join} from 'node:path';
import {fields,helpers} from '../reference.mjs';
const root=fileURLToPath(new URL('../dist/',import.meta.url));
const routes=['/','/get-started/','/setups/','/handbook/','/reference/'];
const pages=new Map(routes.map(r=>[r,readFileSync(join(root,r,'index.html'),'utf8')]));
test('every local page, section and asset link resolves in portable output',()=>{
 for(const [route,html] of pages){
  const ids=[...html.matchAll(/\bid="([^"]+)"/g)].map(m=>m[1]);
  assert.equal(new Set(ids).size,ids.length,`${route}: duplicate anchors`);
  for(const m of html.matchAll(/(?:href|src)="([^"]+)"/g)){
   const href=m[1].replaceAll('&amp;','&');if(!href.startsWith('/')&&!href.startsWith('#'))continue;
   const url=new URL(href,'https://example.test'+route);
   if(pages.has(url.pathname)){
    if(url.hash)assert.ok(pages.get(url.pathname).includes(`id="${url.hash.slice(1)}"`),`${route}: broken ${href}`);
   }else assert.ok(existsSync(join(root,url.pathname)),`${route}: missing ${href}`);
  }
 }
});
test('reference covers current YAML schema recursively and all registered helpers',()=>{
 const go=readFileSync(new URL('../../config/config.go',import.meta.url),'utf8');
 const structs=new Map([...go.matchAll(/type (\w+) struct \{([\s\S]*?)\n\}/g)].map(m=>[m[1],m[2]]));
 const expected=[];
 function visit(name,prefix=''){
  for(const line of structs.get(name).split('\n')){
   const m=line.match(/^(.*?)`yaml:"([^"]+)"`/);if(!m)continue;
   const type=m[1].trim().split(/\s+/).at(-1),key=prefix+m[2];expected.push(key);
   const child=type.replace('[]','');if(structs.has(child))visit(child,key+(type.startsWith('[]')?'[].':'.'));
  }
 }
 visit('Config');assert.deepEqual(fields.map(f=>f.key).sort(),expected.sort());
 const funcs=readFileSync(new URL('../../tmpl/funcs.go',import.meta.url),'utf8').split('func removeLines')[0];
 assert.deepEqual(helpers.map(h=>h[0]).sort(),[...funcs.matchAll(/"(\w+)":/g)].map(m=>m[1]).sort());
});
test('tool concepts and advanced explanations are readable without publishing research workspaces',()=>{
 for(const id of ['execution','compiled','interpreted','project-builds','source-shaping','portability'])assert.ok(pages.get('/setups/').includes(`id="${id}"`));
 for(const html of pages.values())assert.doesNotMatch(html,/cp_x|source workspace|source branch|twelve (?:configurations|setups)|gbuild|leanbuild/);
 for(const path of ['recipe-fixtures.json','transformation-fixtures.json','examples'])assert.equal(existsSync(join(root,path)),false);
 assert.ok(pages.get('/handbook/').includes('id="adapt-your-language"'));
 assert.ok(pages.get('/handbook/').includes('Prepared source · normalized whitespace'));
 assert.match(pages.get('/'),/Your language\.<br>Your editor\./);
 assert.doesNotMatch(pages.get('/'),/class="language-line">C\+\+/);
 for(const html of pages.values()){assert.match(html,/MonaspaceNeonVar\.woff2/);for(const route of routes.slice(1))assert.ok(html.includes(`href="${route}"`));}
});
test('production pages identify their public URLs and exclude test fixtures from upload',()=>{
 const origin='https://eitri.veryshyjelly.in';
 for(const [route,html] of pages){
  assert.ok(html.includes(`<link rel="canonical" href="${origin}${route}">`));
  assert.ok(html.includes(`<meta property="og:url" content="${origin}${route}">`));
  assert.ok(html.includes(`<meta property="og:image" content="${origin}/assets/social.svg">`));
 }
 assert.match(readFileSync(join(root,'.assetsignore'),'utf8'),/^fixtures\.json$/m);
 assert.match(readFileSync(join(root,'404.html'),'utf8'),/name="robots" content="noindex"/);
 const sitemap=readFileSync(join(root,'sitemap.xml'),'utf8');
 for(const route of routes)assert.ok(sitemap.includes(`<loc>${origin}${route}</loc>`));
});
