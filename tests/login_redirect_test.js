const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');

const scriptPath = path.join(__dirname, '../internal/resources/static/js/login.js');
const loginScript = fs.readFileSync(scriptPath, 'utf8');

async function submitLogin(search) {
    let submit;
    const button = {textContent: 'Login', disabled: false};
    const form = {
        addEventListener: (event, handler) => { if (event === 'submit') submit = handler; },
        querySelector: () => button,
    };
    const elements = {
        loginForm: form,
        loginError: {dataset: {errorMessage: 'Invalid credentials'}, style: {}},
        username: {value: 'viewer'},
        password: {value: 'password'},
        keepLoggedIn: {checked: false},
    };
    const document = {
        getElementById: id => elements[id],
        addEventListener: (event, handler) => { if (event === 'DOMContentLoaded') handler(); },
    };
    const window = {location: {origin: 'https://wiki.example', search, href: ''}};
    let fetchCalls = 0;
    const fetch = async (url, options) => {
        fetchCalls++;
        assert.equal(url, '/api/login');
        assert.equal(options.method, 'POST');
        return {ok: true};
    };
    vm.runInNewContext(loginScript, {document, window, fetch, URL, URLSearchParams, console}, {filename: scriptPath});
    assert.equal(typeof submit, 'function');
    let prevented = false;
    await submit({preventDefault() { prevented = true; }});
    assert.ok(prevented);
    assert.equal(fetchCalls, 1);
    return window.location.href;
}

test('successful login stays on the wiki origin', async t => {
    const cases = [
        ['protocol-relative', '?redirect=//evil.example', '/'],
        ['encoded protocol-relative', '?redirect=%2F%2Fevil.example%2Fphish', '/'],
        ['backslash authority', '?redirect=%2F%5Cevil.example', '/'],
        ['control characters', '?redirect=%2F%09%2Fevil.example', '/'],
        ['normalized double slash', '?redirect=%2F%252e%2F%2Fevil.example', '/'],
        ['absolute external URL', '?redirect=https%3A%2F%2Fevil.example', '/'],
        ['unsupported scheme', '?redirect=javascript%3Aalert(1)', '/'],
        ['missing', '', '/'],
        ['empty', '?redirect=', '/'],
        ['local path with query and fragment', '?redirect=%2Ffinance%2Freport%3Fview%3Dfull%23summary', '/finance/report?view=full#summary'],
        ['normalized local path', '?redirect=%2Fa%2F..%2Fb', '/b'],
    ];
    for (const [name, search, expected] of cases) {
        await t.test(name, async () => {
            const destination = await submitLogin(search);
            assert.equal(destination, expected);
            assert.equal(new URL(destination, 'https://wiki.example').origin, 'https://wiki.example');
        });
    }
});
