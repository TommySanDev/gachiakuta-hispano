import fs from 'fs';
import path from 'path';

const imageDirs = [
  { dir: 'src/assets/chars', alias: 'char' },
  { dir: 'src/assets/vi', alias: 'vi' },
  { dir: 'src/assets/chapters', alias: 'chapter' },
];

const outFile = 'src/lib/image-map.ts';

const build = async () => {
  let imports: string[] = [];
  let mapEntries: string[] = [];

  for (const { dir, alias } of imageDirs) {
    if (!fs.existsSync(dir)) {
      console.warn(`⚠️  Skipping missing directory: ${dir}`);
      continue;
    }

    const files = fs
      .readdirSync(dir)
      .filter(f => !f.startsWith('.') && /\.(jpe?g|webp|png)$/i.test(f))
      .sort();

    for (const file of files) {
      const varName = `${alias}_${file.replace(/\W+/g, '_')}`;
      const importPath = `../assets/${path.basename(dir)}/${file}`;
      imports.push(`import ${varName} from '${importPath}';`);
      mapEntries.push(`  '${file}': ${varName},`);
    }
  }

  const final = `// 🚨 AUTO-GENERATED FILE. DO NOT EDIT MANUALLY.
// Run \`pnpm run generate-image-map\` to regenerate.

${imports.join('\n')}

export const imageMap: Record<string, ImageMetadata> = {
${mapEntries.join('\n')}
};
`;

  fs.writeFileSync(outFile, final);
  console.log(`✅ Generated ${outFile}`);
};

build();

