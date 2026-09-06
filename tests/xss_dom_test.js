const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

class FakeTextNode {
    constructor(value) {
        this.textContent = String(value);
    }
}

class FakeElement {
    constructor(tagName) {
        this.tagName = tagName;
        this.childNodes = [];
        this.className = '';
        this.attributes = {};
    }

    set textContent(value) {
        this.childNodes = [new FakeTextNode(value)];
    }

    get textContent() {
        return this.childNodes.map(child => child.textContent).join('');
    }

    get innerHTML() {
        return this.textContent.replace(/[&<>"']/g, character => ({
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#39;',
        })[character]);
    }

    appendChild(child) {
        this.childNodes.push(child);
        return child;
    }

    replaceChildren(...children) {
        this.childNodes = children;
    }

    setAttribute(name, value) {
        this.attributes[name] = String(value);
    }
}

global.window = {
    location: {
        origin: 'https://wiki.example',
        pathname: '/docs/security',
    },
};
global.document = {
    createElement: tagName => new FakeElement(tagName),
    createTextNode: value => new FakeTextNode(value),
};

const utilitiesPath = path.join(__dirname, '..', 'internal', 'resources', 'static', 'js', 'utilities.js');
vm.runInThisContext(fs.readFileSync(utilitiesPath, 'utf8'), {filename: utilitiesPath});
global.WikiDOM = window.WikiDOM;

test('user and API values remain text in DOM helpers', () => {
    const payloads = [
        '<img src=x onerror=globalThis.pwned=true>',
        '\"><svg onload=globalThis.pwned=true>',
        '<a href=javascript:globalThis.pwned=true>click</a>',
    ];

    for (const payload of payloads) {
        const element = window.WikiDOM.element('span', 'filename', payload);
        assert.equal(element.textContent, payload);
        assert.equal(element.childNodes.length, 1);

        const container = new FakeElement('div');
        window.WikiDOM.message(container, 'error-message', payload);
        assert.equal(container.textContent, payload);
        assert.equal(container.childNodes.length, 1);
    }
    assert.equal(globalThis.pwned, undefined);
});

test('highlighting creates text and known span nodes only', () => {
    const payload = '<img src=x onerror=alert(1)> report';
    const container = new FakeElement('div');
    window.WikiDOM.appendHighlightedText(container, payload, /(report)/gi);
    assert.equal(container.textContent, payload);
    assert.equal(container.childNodes[1].tagName, 'span');
    assert.equal(container.childNodes[1].className, 'search-result-highlight');
});

test('API-provided links are restricted to this origin', () => {
    assert.equal(window.WikiDOM.localURL('/docs/report?q=1#part'), '/docs/report?q=1#part');
    for (const payload of ['javascript:alert(1)', 'data:text/html,<script>alert(1)</script>', 'https://evil.example/']) {
        assert.equal(window.WikiDOM.localURL(payload), '#');
    }
});

test('browser Markdown rejects active link and image schemes', () => {
    assert.equal(window.WikiDOM.markdownURL('/docs/report'), '/docs/report');
    assert.equal(window.WikiDOM.markdownURL('https://example.com/report'), 'https://example.com/report');
    assert.equal(window.WikiDOM.markdownURL('mailto:security@example.com'), 'mailto:security@example.com');
    assert.equal(window.WikiDOM.markdownURL('javascript:alert(1)'), null);
    assert.equal(window.WikiDOM.markdownURL('data:text/html,<script>alert(1)</script>'), null);
    assert.equal(window.WikiDOM.markdownURL('mailto:security@example.com', true), null);
});

test('Kanban browser rendering keeps raw HTML inert and blocks active URLs', () => {
    const kanbanUIPath = path.join(__dirname, '..', 'internal', 'resources', 'static', 'js', 'kanban-ui.js');
    const kanbanSource = fs.readFileSync(kanbanUIPath, 'utf8') + '\nglobalThis.KanbanUIManagerForTest = KanbanUIManager;';
    vm.runInThisContext(kanbanSource, {filename: kanbanUIPath});
    const manager = new globalThis.KanbanUIManagerForTest(null);

    const rendered = manager.processMarkdown(
        '<svg onload="globalThis.pwned=true"></svg> [click](javascript:globalThis.pwned=true) ![x](data:image/svg+xml,<svg onload=alert(1)>)'
    );
    assert.doesNotMatch(rendered, /<svg|href="javascript:|src="data:/i);
    assert.match(rendered, /&lt;svg/);
    assert.equal(globalThis.pwned, undefined);
});
