import commonjs from 'rollup-plugin-commonjs';
import json from 'rollup-plugin-json';
import progress from 'rollup-plugin-progress';
import resolve from 'rollup-plugin-node-resolve';
import typescript from 'rollup-plugin-typescript2';

const tsconfig = process.env.NODE_ENV !== 'production' ? 'tsconfig.dev.json' : 'tsconfig.json';

const plugins = [
    typescript({ tsconfig }),
    resolve(),
    json(),
    commonjs(),
    progress(),
];

export default [
    {
        input: './src/chat.ts',
        output: {
            file: './lib/chat.js',
            format: 'umd',
            name: "Chat",
            sourcemap: process.env.NODE_ENV !== 'production',
        },
        plugins,
    },
];
