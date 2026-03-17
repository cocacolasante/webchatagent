import * as esbuild from 'esbuild';
import { readFileSync } from 'fs';
import { gzipSync } from 'zlib';

const watch = process.argv.includes('--watch');

// Plugin to inline CSS files as JS string exports
const inlineCSSPlugin = {
  name: 'inline-css',
  setup(build) {
    build.onLoad({ filter: /\.css$/ }, async (args) => {
      const css = readFileSync(args.path, 'utf8');
      return {
        contents: `export default ${JSON.stringify(css)};`,
        loader: 'js',
      };
    });
  },
};

const buildOptions = {
  entryPoints: ['src/widget.ts'],
  bundle: true,
  outfile: 'dist/widget.js',
  format: 'iife',
  target: ['es2017'],
  minify: !watch,
  sourcemap: watch ? 'inline' : false,
  platform: 'browser',
  plugins: [inlineCSSPlugin],
  define: {
    'process.env.NODE_ENV': watch ? '"development"' : '"production"',
  },
};

if (watch) {
  const ctx = await esbuild.context(buildOptions);
  await ctx.watch();
  console.log('Watching for changes...');
} else {
  const result = await esbuild.build(buildOptions);

  // Check bundle size
  const data = readFileSync('dist/widget.js');
  const gzipped = gzipSync(data);
  const maxKB = parseInt(process.env.WIDGET_MAX_SIZE_KB || '50');
  const gzippedKB = (gzipped.length / 1024).toFixed(1);
  const rawKB = (data.length / 1024).toFixed(1);

  console.log(`Widget built: ${rawKB}KB raw, ${gzippedKB}KB gzipped`);

  if (gzipped.length > maxKB * 1024) {
    console.error(`ERROR: Widget too large: ${gzippedKB}KB gzipped (max ${maxKB}KB)`);
    process.exit(1);
  }

  console.log(`✓ Widget size OK (${gzippedKB}KB / ${maxKB}KB limit)`);
}
