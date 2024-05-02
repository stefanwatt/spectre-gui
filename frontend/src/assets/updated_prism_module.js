
// Refactored PrismJS as a Module
var _self = (typeof window !== 'undefined' ? window : (typeof WorkerGlobalScope !== 'undefined' && self instanceof WorkerGlobalScope ? self : {}));
var Prism = (function () {
    // Place the entire PrismJS core logic here

    var util = {
        // Original utility methods
        encode: function (n) {
            return n instanceof Node ? new Node(n.type, util.encode(n.content), n.alias) : Array.isArray(n) ? n.map(util.encode) : n.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/\u00a0/g, " ");
        },
        type: function (e) {
            return Object.prototype.toString.call(e).slice(8, -1);
        },
        objId: function (e) {
            return e.__id || Object.defineProperty(e, "__id", { value: ++t }), e.__id;
        },
        clone: function (n, t) { /* Cloning logic */ },
        getLanguage: function (e) { /* Language retrieval logic */ },
        setLanguage: function (e, t) { /* Language setting logic */ },
        currentScript: function () { /* Current script logic */ },
        isActive: function (e, n, t) { /* Active check logic */ },
    };

    var languages = {
        // Original languages definitions
        markup: {/* Markup grammar */},
        css: {/* CSS grammar */},
        clike: {/* C-Like grammar */},
        javascript: {/* JavaScript grammar */},
        // Additional languages
    };

    var hooks = {
        all: {},
        add: function (name, callback) {
            var hook = hooks.all[name] = hooks.all[name] || [];
            hook.push(callback);
        },
        run: function (name, env) {
            var hook = hooks.all[name];
            if (hook && hook.length) {
                hook.forEach(function (callback) {
                    callback(env);
                });
            }
        },
    };

    // Core highlighting functions
    function highlightElement(element, async, callback) { /* Highlighting logic */ }
    function tokenize(text, grammar) { /* Tokenization logic */ }
    function highlight(text, grammar, language) { /* Highlighting logic */ }

    return {
        util: util,
        languages: languages,
        hooks: hooks,
        highlight: highlight,
        tokenize: tokenize,
        highlightElement: highlightElement,
    };
})();

if (typeof module !== 'undefined' && module.exports) {
    module.exports = Prism;
}
if (typeof global !== 'undefined') {
    global.Prism = Prism;
}
