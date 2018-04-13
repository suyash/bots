import commonjs from 'rollup-plugin-commonjs';
import json from 'rollup-plugin-json';
import progress from 'rollup-plugin-progress';
import resolve from 'rollup-plugin-node-resolve';
import typescript from 'rollup-plugin-typescript2';

const tsconfig = process.env.NODE_ENV !== 'production' ? 'tsconfig.dev.json' : 'tsconfig.json';

const plugins = [
    typescript({ tsconfig }),
    resolve({ browser: true }),
    json(),
    commonjs(),
    progress(),
];

const modules = [
    {
        input: "./src/echo.ts",
        output: "./lib/echo.js",
        name: "Echo",
    },
    {
        input: "./src/threads.ts",
        output: "./lib/threads.js",
        name: "Threads",
    },
    {
        input: "./src/attachments.ts",
        output: "./lib/attachments.js",
        name: "Attachments",
    },
    {
        input: "./src/conversations.ts",
        output: "./lib/conversations.js",
        name: "Conversations",
    },
    {
        input: "./src/redis.ts",
        output: "./lib/redis.js",
        name: "Redis",
    },
];

export default modules.map(m => ({
    input: m.input,
    output: {
        file: m.output,
        format: 'umd',
        name: m.name,
        sourcemap: process.env.NODE_ENV !== 'production',
    },
    plugins,
}));
