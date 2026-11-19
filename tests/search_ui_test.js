const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

const staticDir = path.join(__dirname, '../internal/resources/static/js');
const utilitiesSource = fs.readFileSync(path.join(staticDir, 'utilities.js'), 'utf8');
const searchSource = fs.readFileSync(path.join(staticDir, 'search.js'), 'utf8');
const basePathSource = fs.readFileSync(path.join(staticDir, 'base-path.js'), 'utf8');

class Element {
    constructor(tagName = 'div') {
        this.tagName = tagName;
        this.childNodes = [];
        this.handlers = {};
        this.classes = new Set();
        this.classList = {
            add: name => this.classes.add(name),
            remove: name => this.classes.delete(name),
            contains: name => this.classes.has(name),
        };
        this.value = '';
        this.href = '';
    }
    set textContent(value) { this.childNodes = [new TextNode(value)]; }
    get textContent() { return this.childNodes.map(child => child.textContent).join(''); }
    set innerHTML(_) { throw new Error('untrusted HTML was parsed'); }
    append(...children) { this.childNodes.push(...children); }
    appendChild(child) { this.childNodes.push(child); return child; }
    replaceChildren(...children) { this.childNodes = children; }
    addEventListener(name, handler) { this.handlers[name] = handler; }
    dispatch(name, event = {}) { return this.handlers[name]?.(event); }
}

class TextNode {
    constructor(text) { this.textContent = String(text); }
}

function setup(pageMode, fetch, basePath = '') {
    const selectors = new Map();
    for (const selector of ['.search-box', '.search-form']) {
        selectors.set(selector, new Element());
    }
    if (pageMode) {
        for (const selector of ['.search-page', '.search-page-form', '#search-page-query', '.search-page-results']) {
            selectors.set(selector, new Element());
        }
    }
    const documentHandlers = {};
    const windowHandlers = {};
    selectors.set('meta[name="wiki-base-path"]', {content: basePath});
    const document = {
        querySelector: selector => selectors.get(selector) || null,
        createElement: tag => new Element(tag),
        createTextNode: text => new TextNode(text),
        createDocumentFragment: () => new Element('fragment'),
        addEventListener: (name, handler) => { documentHandlers[name] = handler; },
    };
    const initial = 'https://wiki.example' + basePath + (pageMode ? '/?search=old' : '/docs/start');
    const location = {origin: 'https://wiki.example'};
    const setLocation = target => {
        const url = new URL(target, location.origin);
        Object.assign(location, {href: url.href, pathname: url.pathname, search: url.search});
    };
    setLocation(initial);
    location.assign = setLocation;
    const historyCalls = [];
    const window = {
        location,
        history: {
            pushState: (_state, _title, target) => { historyCalls.push(target); setLocation(target); },
            replaceState: (_state, _title, target) => { historyCalls.push(target); setLocation(target); },
        },
        addEventListener: (name, handler) => { windowHandlers[name] = handler; },
    };
    const timers = [];
    const context = {window, document, fetch, URL, URLSearchParams, AbortController, console,
        setTimeout: callback => { timers.push(callback); return timers.length - 1; },
        clearTimeout: id => { timers[id] = null; }};
    vm.runInNewContext(basePathSource, context, {filename: 'base-path.js'});
    context.wikiURL = window.wikiURL;
    context.wikiPathname = window.wikiPathname;
    vm.runInNewContext(utilitiesSource, context, {filename: 'utilities.js'});
    context.WikiDOM = window.WikiDOM;
    vm.runInNewContext(searchSource, context, {filename: 'search.js'});
    documentHandlers.DOMContentLoaded();
    return {
        get: selector => selectors.get(selector),
        location, setLocation, historyCalls,
        popstate: () => windowHandlers.popstate(),
        runTimers: () => { for (const timer of timers.splice(0)) if (timer) timer(); },
    };
}

async function flush() {
    for (let i = 0; i < 5; i++) await Promise.resolve();
}

test('typing in sidebar opens the search page without Enter', async () => {
    const app = setup(false, async () => { throw new Error('navigation should load the page'); });
    const box = app.get('.search-box');
    box.value = '"report" + café';
    box.dispatch('input', {target: box});
    assert.equal(app.location.pathname, '/docs/start');
    assert.equal(app.historyCalls.length, 0);
    app.runTimers();
    assert.equal(app.location.pathname, '/');
    assert.equal(new URL(app.location.href).searchParams.get('search'), '"report" + café');
    let prevented = false;
    app.get('.search-form').dispatch('submit', {preventDefault() { prevented = true; }});
    assert.equal(prevented, false, 'native GET form should navigate from a document');
});

test('search-page form updates URL and popstate reloads results', async () => {
    const requests = [];
    const app = setup(true, async target => {
        requests.push(target);
        return {ok: true, json: async () => [{title: target, path: '/public', excerpt: 'safe'}]};
    });
    const input = app.get('#search-page-query');
    input.value = '"café + notes"';
    let prevented = false;
    app.get('.search-page-form').dispatch('submit', {preventDefault() { prevented = true; }});
    await flush();
    assert.ok(prevented);
    assert.equal(app.historyCalls.length, 1);
    assert.equal(new URL(app.location.href).searchParams.get('search'), '"café + notes"');
    assert.equal(new URL(requests[0], app.location.origin).searchParams.get('q'), '"café + notes"');
    assert.match(app.get('.search-page-results').textContent, /safe/);
    app.setLocation('https://wiki.example/?search=old');
    app.popstate();
    await flush();
    assert.equal(input.value, 'old');
    assert.equal(app.get('.search-box').value, 'old');
    assert.equal(new URL(requests[1], app.location.origin).searchParams.get('q'), 'old');
});

test('late page responses cannot replace newer search results', async () => {
    const pending = [];
    const app = setup(true, target => new Promise(resolve => pending.push({target, resolve})));
    const input = app.get('#search-page-query');
    const submit = query => {
        input.value = query;
        app.get('.search-page-form').dispatch('submit', {preventDefault() {}});
    };
    submit('first');
    submit('second');
    assert.equal(pending.length, 2);
    pending[1].resolve({ok: true, json: async () => [{title: 'SECOND', path: '/second', excerpt: 'new'}]});
    await flush();
    pending[0].resolve({ok: true, json: async () => [{title: 'FIRST', path: '/first', excerpt: 'old'}]});
    await flush();
    assert.match(app.get('.search-page-results').textContent, /SECOND/);
    assert.doesNotMatch(app.get('.search-page-results').textContent, /FIRST/);
});

test('live search keeps titles and excerpts as text and restricts links', async () => {
    const app = setup(true, async () => ({ok: true, json: async () => [{
        title: '<img src=x onerror=alert(1)> report',
        path: 'javascript:alert(1)',
        excerpt: '<svg onload=alert(1)>',
    }]}));
    const box = app.get('.search-box');
    box.value = 'report';
    box.dispatch('input', {target: box});
    app.runTimers();
    await flush();
    const results = app.get('.search-page-results');
    assert.match(results.textContent, /<img src=x onerror=alert\(1\)> report/);
    assert.match(results.textContent, /<svg onload=alert\(1\)>/);
    assert.equal(results.childNodes[0].childNodes[0].childNodes[0].href, '#');
});

test('typing on the results page updates URL and ignores older responses', async () => {
    const pending = [];
    const app = setup(true, target => new Promise(resolve => pending.push({target, resolve})));
    const box = app.get('.search-box');
    box.value = 'first';
    box.dispatch('input', {target: box});
    app.runTimers();
    box.value = 'second';
    box.dispatch('input', {target: box});
    app.runTimers();
    assert.equal(pending.length, 2);
    pending[1].resolve({ok: true, json: async () => [{title: 'SECOND', path: '/second', excerpt: 'new'}]});
    await flush();
    pending[0].resolve({ok: true, json: async () => [{title: 'FIRST', path: '/first', excerpt: 'old'}]});
    await flush();
    assert.match(app.get('.search-page-results').textContent, /SECOND/);
    assert.doesNotMatch(app.get('.search-page-results').textContent, /FIRST/);
    assert.equal(new URL(app.location.href).searchParams.get('search'), 'second');

    let prevented = false;
    app.get('.search-form').dispatch('submit', {preventDefault() { prevented = true; }});
    assert.ok(prevented);
    assert.equal(new URL(app.location.href).searchParams.get('search'), 'second');
    assert.equal(app.historyCalls.length, 2);
    assert.equal(new URL(pending[2].target, app.location.origin).searchParams.get('q'), 'second');
});

test('search under a base path keeps navigation, requests, and result links inside it', async () => {
    const requests = [];
    const app = setup(false, async target => {
        requests.push(target);
        return {ok: true, json: async () => []};
    }, '/wiki');
    const box = app.get('.search-box');
    box.value = 'notes';
    box.dispatch('input', {target: box});
    app.runTimers();
    assert.equal(app.location.pathname, '/wiki/');
    assert.equal(new URL(app.location.href).searchParams.get('search'), 'notes');

    const page = setup(true, async target => {
        requests.push(target);
        return {ok: true, json: async () => [{title: 'Guide', path: '/docs/guide', excerpt: 'safe'}]};
    }, '/wiki');
    const input = page.get('#search-page-query');
    input.value = 'guide';
    page.get('.search-page-form').dispatch('submit', {preventDefault() {}});
    await flush();
    assert.equal(page.location.pathname, '/wiki/');
    assert.equal(new URL(requests.at(-1), page.location.origin).pathname, '/wiki/api/search');
    const result = page.get('.search-page-results').childNodes[0].childNodes[0].childNodes[0];
    assert.equal(result.href, '/wiki/docs/guide');
});
