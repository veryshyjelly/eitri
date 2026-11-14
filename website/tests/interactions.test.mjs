import {test} from 'node:test';
import assert from 'node:assert/strict';
import vm from 'node:vm';
import {readFileSync} from 'node:fs';
const source=readFileSync(new URL('../app.js',import.meta.url),'utf8').replace(/^import[^\n]*\n/,'');
function element(){return {hidden:true,textContent:'',attrs:{},events:{},setAttribute(k,v){this.attrs[k]=v},addEventListener(k,v){this.events[k]=v}}}
function setup({storageFails=false,clipboardFails=false,media=false}={}){
 const theme=element(),copy=element(),status=element(),play=media?element():null;
 const code={textContent:'go version\n'},block={querySelector:s=>s==='code'?code:status};copy.closest=()=>block;
 let stored={},video=null;
 const doc={documentElement:{dataset:{theme:'dark'}},querySelector:s=>({'#theme':theme,'#lang':null,'#play':play,'#demo':{replaceChildren(v){video=v}}}[s]),querySelectorAll:s=>s==='.copy'?[copy]:[],createElement:tag=>({...element(),tag,append(){},focus(){},play(){return Promise.resolve()}})};
 const localStorage={setItem(k,v){if(storageFails)throw Error('denied');stored[k]=v}};
 let copied;
 vm.runInNewContext(source,{document:doc,localStorage,navigator:{clipboard:{async writeText(text){if(clipboardFails)throw Error('denied');copied=text}}}});
 return {theme,copy,status,play,doc,stored,get copied(){return copied},get video(){return video}};
}
test('theme toggle persists explicit choice and survives blocked storage',()=>{for(const storageFails of [false,true]){const s=setup({storageFails});s.theme.events.click();assert.equal(s.doc.documentElement.dataset.theme,'light');assert.equal(s.theme.attrs['aria-label'],'Switch to dark mode');if(!storageFails)assert.equal(s.stored['eitri-theme'],'light')}});
test('copy feedback only claims success after clipboard success',async()=>{const s=setup();await s.copy.events.click();assert.equal(s.copied,'go version\n');assert.equal(s.status.textContent,'Copied to clipboard.');const denied=setup({clipboardFails:true});await denied.copy.events.click();assert.match(denied.status.textContent,/Select the text/);assert.notEqual(denied.status.textContent,'Copied to clipboard.')});
test('media source is attached only after deliberate playback',()=>{const s=setup({media:true});assert.equal(s.video,null);s.play.events.click();assert.equal(s.video.src,'/assets/workflow.mp4');assert.equal(s.video.controls,true)});
test('early preference read accepts valid values and handles storage failures',()=>{const html=readFileSync(new URL('../dist/index.html',import.meta.url),'utf8'),early=html.match(/<script>(.*?)<\/script>/s)[1];for(const value of ['light','dark','invalid',null]){const document={documentElement:{dataset:{theme:'dark'}}};vm.runInNewContext(early,{document,localStorage:{getItem:()=>value}});assert.equal(document.documentElement.dataset.theme,value==='light'?'light':'dark')}const document={documentElement:{dataset:{theme:'dark'}}};vm.runInNewContext(early,{document,localStorage:{getItem(){throw Error()}}});assert.equal(document.documentElement.dataset.theme,'dark')});
