const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

class FakeClassList {
    constructor(...names) {
        this.names = new Set(names);
    }

    add(name) {
        this.names.add(name);
    }

    remove(name) {
        this.names.delete(name);
    }

    contains(name) {
        return this.names.has(name);
    }

    toggle(name, force) {
        const enabled = force === undefined ? !this.contains(name) : Boolean(force);
        if (enabled) {
            this.add(name);
        } else {
            this.remove(name);
        }
        return enabled;
    }
}

class FakeElement {
    constructor(className, dataset = {}) {
        this.classList = new FakeClassList(className);
        this.dataset = dataset;
        this.attributes = {};
        this.listeners = {};
        this.style = {};
        this.children = [];
        this.focusCount = 0;
    }

    setAttribute(name, value) {
        this.attributes[name] = String(value);
    }

    toggleAttribute(name, force) {
        if (force) {
            this.attributes[name] = '';
        } else {
            delete this.attributes[name];
        }
    }

    addEventListener(name, callback) {
        this.listeners[name] = callback;
    }

    querySelector(selector) {
        return selector === '.chapter-links-body' ? this.body : null;
    }

    contains(element) {
        return this.children.includes(element);
    }

    setPointerCapture() {}

    getBoundingClientRect() {
        return {top: Number.parseInt(this.style.top || '100', 10)};
    }
}

function initializeChapterLinks(labels, initiallyRetracted, coarsePointer = false) {
    const root = {classList: new FakeClassList()};
    root.classList.toggle('chapter-links-retracted', initiallyRetracted);

    const body = new FakeElement('chapter-links-body');
    const focusedLink = new FakeElement('chapter-link');
    const panel = new FakeElement('chapter-links');
    panel.body = body;
    panel.children.push(body, focusedLink);

    const toggle = new FakeElement('chapter-links-toggle', {
        labelExpand: labels.expand,
        labelRetract: labels.retract,
    });

    let readyHandler;
    const document = {
        documentElement: root,
        activeElement: null,
        querySelectorAll: () => [],
        querySelector: selector => ({
            '.chapter-links': panel,
            '.chapter-links-toggle': toggle,
        })[selector] || null,
        addEventListener: (name, callback) => {
            if (name === 'DOMContentLoaded') readyHandler = callback;
        },
    };
    toggle.focus = () => {
        toggle.focusCount += 1;
        document.activeElement = toggle;
    };

    const stored = {};
    const context = {
        document,
        window: {
            addEventListener: () => {},
            matchMedia: () => ({matches: coarsePointer}),
            innerHeight: 800,
        },
        sessionStorage: {
            getItem: key => stored[key] ?? null,
            setItem: (key, value) => {
                stored[key] = String(value);
            },
        },
        requestAnimationFrame: callback => callback(),
    };

    const controllerPath = path.join(__dirname, '..', 'internal', 'resources', 'static', 'js', 'markdown-extensions.js');
    vm.runInNewContext(fs.readFileSync(controllerPath, 'utf8'), context, {filename: controllerPath});
    readyHandler();

    return {body, document, focusedLink, panel, root, stored, toggle};
}

test('chapter links synchronize translated accessible state', () => {
    const labels = {
        expand: 'Expand translated',
        retract: 'Retract translated',
    };
    const state = initializeChapterLinks(labels, true);

    assert.equal(state.toggle.attributes['aria-expanded'], 'false');
    assert.equal(state.toggle.attributes['aria-label'], labels.expand);
    assert.equal(state.panel.attributes['aria-hidden'], 'true');
    assert.equal(state.body.attributes['aria-hidden'], 'true');
    assert.ok(Object.hasOwn(state.panel.attributes, 'inert'));

    state.toggle.listeners.click();
    assert.equal(state.toggle.attributes['aria-expanded'], 'true');
    assert.equal(state.toggle.attributes['aria-label'], labels.retract);
    assert.equal(state.panel.attributes['aria-hidden'], 'false');
    assert.equal(state.body.attributes['aria-hidden'], 'false');
    assert.ok(!Object.hasOwn(state.panel.attributes, 'inert'));

    state.document.activeElement = state.focusedLink;
    state.toggle.listeners.click();
    assert.equal(state.toggle.focusCount, 1);
    assert.equal(state.document.activeElement, state.toggle);
    assert.equal(state.stored['chapter-links-retracted'], 'true');
});

test('chapter links preserve non-English labels from template data', () => {
    const labels = {
        expand: 'توسيع روابط الفصول',
        retract: 'طي روابط الفصول',
    };
    const state = initializeChapterLinks(labels, false);

    assert.equal(state.toggle.attributes['aria-label'], labels.retract);
    state.toggle.listeners.click();
    assert.equal(state.toggle.attributes['aria-label'], labels.expand);
});

test('mobile drag does not trigger a trailing panel toggle', () => {
    const state = initializeChapterLinks({
        expand: 'Expand',
        retract: 'Retract',
    }, false, true);

    state.toggle.listeners.pointerdown({
        pointerType: 'touch',
        pointerId: 1,
        clientY: 100,
    });
    state.toggle.listeners.pointermove({pointerId: 1, clientY: 150});
    state.toggle.listeners.pointerup({pointerId: 1});
    state.toggle.listeners.click();

    assert.equal(state.toggle.style.top, '150px');
    assert.equal(state.stored['chapter-links-toggle-top'], '150');
    assert.equal(state.toggle.attributes['aria-expanded'], 'true');
});
