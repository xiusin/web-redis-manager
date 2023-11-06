// http://eslint.org/docs/user-guide/configuring

module.exports = {
    root: true,
    parser: 'babel-eslint',
    parserOptions: {
        sourceType: 'module'
    },
    env: {
        browser: true,
    },
    globals: {
        "Vue": true,
        "jQuery": true
    },
    // https://github.com/feross/standard/blob/master/RULES.md#javascript-standard-style
    extends: 'standard',
    // required to lint *.vue files
    plugins: [
        'html'
    ],
    // add your custom rules here
    'rules': {
        // disallow check for semicolon
        'semi': 0,
        // disable indent rule
        "indent": "off",
        // allow paren-less arrow functions
        'arrow-parens': 0,
        'space-before-function-paren': ['off', 'never'],
        // allow async-await
        'generator-star-spacing': 0,
        // allow debugger during development
        'no-debugger': process.env.NODE_ENV === 'production' ? 2 : 0,
        'no-unused-vars': 0
    }
}
