import { toPng } from 'html-to-image';

export async function exportElementAsPng(element: HTMLElement, filename: string): Promise<void> {
  const dataUrl = await toPng(element, {
    backgroundColor: '#0a0e17',
    pixelRatio: 2,
    cacheBust: true,
  });
  const link = document.createElement('a');
  link.download = filename;
  link.href = dataUrl;
  link.click();
}
