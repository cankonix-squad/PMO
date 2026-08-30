'use client';

import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Placeholder from '@tiptap/extension-placeholder';
import {
  Bold,
  Italic,
  List,
  ListOrdered,
  Minus,
  Paperclip,
  Image as ImageIcon,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface RichTextEditorProps {
  /** Controlled HTML content */
  value: string;
  onChange: (html: string) => void;
  placeholder?: string;
  disabled?: boolean;
  /** Called when user clicks the attach-file button */
  onAttachFile?: () => void;
  /** Called when user pastes or picks an image */
  onAttachImage?: (file: File) => void;
  className?: string;
}

export function RichTextEditor({
  value,
  onChange,
  placeholder = 'Tulis di sini...',
  disabled = false,
  onAttachFile,
  onAttachImage,
  className,
}: RichTextEditorProps) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder }),
    ],
    content: value || '',
    editable: !disabled,
    onUpdate({ editor }) {
      // Return empty string when content is just an empty paragraph
      const html = editor.isEmpty ? '' : editor.getHTML();
      onChange(html);
    },
    // Sync external value changes (e.g. when form is reset)
    onCreate({ editor }) {
      if (value && value !== editor.getHTML()) {
        editor.commands.setContent(value, false);
      }
    },
  });

  // Handle paste events that include images
  const handlePaste = (e: React.ClipboardEvent<HTMLDivElement>) => {
    if (!onAttachImage) return;
    const items = Array.from(e.clipboardData?.items ?? []);
    const imageItem = items.find((i) => i.type.startsWith('image/'));
    if (imageItem) {
      e.preventDefault();
      const file = imageItem.getAsFile();
      if (file) onAttachImage(file);
    }
  };

  if (!editor) return null;

  const toolbarBtn = (active: boolean, title: string, onClick: () => void, children: React.ReactNode) => (
    <button
      type="button"
      title={title}
      onMouseDown={(e) => {
        e.preventDefault(); // keep editor focus
        onClick();
      }}
      disabled={disabled}
      className={cn(
        'inline-flex h-7 w-7 items-center justify-center rounded text-sm transition-colors',
        active
          ? 'bg-primary/15 text-primary'
          : 'text-muted-foreground hover:bg-muted hover:text-foreground',
        disabled && 'cursor-not-allowed opacity-40'
      )}
    >
      {children}
    </button>
  );

  return (
    <div
      className={cn(
        'rounded-md border border-input bg-background text-sm shadow-sm',
        disabled && 'opacity-60',
        className
      )}
    >
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-0.5 border-b border-input px-2 py-1.5">
        {toolbarBtn(
          editor.isActive('bold'),
          'Tebal (Ctrl+B)',
          () => editor.chain().focus().toggleBold().run(),
          <Bold className="h-3.5 w-3.5" />
        )}
        {toolbarBtn(
          editor.isActive('italic'),
          'Miring (Ctrl+I)',
          () => editor.chain().focus().toggleItalic().run(),
          <Italic className="h-3.5 w-3.5" />
        )}

        <div className="mx-1 h-4 w-px bg-border" />

        {toolbarBtn(
          editor.isActive('bulletList'),
          'Daftar Butir',
          () => editor.chain().focus().toggleBulletList().run(),
          <List className="h-3.5 w-3.5" />
        )}
        {toolbarBtn(
          editor.isActive('orderedList'),
          'Daftar Bernomor',
          () => editor.chain().focus().toggleOrderedList().run(),
          <ListOrdered className="h-3.5 w-3.5" />
        )}

        <div className="mx-1 h-4 w-px bg-border" />

        {toolbarBtn(
          false,
          'Garis Pemisah',
          () => editor.chain().focus().setHorizontalRule().run(),
          <Minus className="h-3.5 w-3.5" />
        )}

        {onAttachFile &&
          toolbarBtn(false, 'Lampirkan File', onAttachFile, <Paperclip className="h-3.5 w-3.5" />)}

        {onAttachImage && (
          <label
            title="Lampirkan Gambar"
            className={cn(
              'inline-flex h-7 w-7 cursor-pointer items-center justify-center rounded text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground',
              disabled && 'cursor-not-allowed opacity-40 pointer-events-none'
            )}
          >
            <ImageIcon className="h-3.5 w-3.5" />
            <input
              type="file"
              accept="image/*"
              className="sr-only"
              disabled={disabled}
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) onAttachImage(file);
                e.target.value = ''; // allow re-picking same file
              }}
            />
          </label>
        )}
      </div>

      {/* Editor content area */}
      <div onPaste={handlePaste}>
        <EditorContent
          editor={editor}
          className={cn(
            'prose prose-sm max-w-none px-3 py-2',
            '[&_.ProseMirror]:min-h-[80px] [&_.ProseMirror]:outline-none',
            '[&_.ProseMirror_p.is-editor-empty:first-child::before]:pointer-events-none',
            '[&_.ProseMirror_p.is-editor-empty:first-child::before]:float-left',
            '[&_.ProseMirror_p.is-editor-empty:first-child::before]:h-0',
            '[&_.ProseMirror_p.is-editor-empty:first-child::before]:text-muted-foreground',
            '[&_.ProseMirror_p.is-editor-empty:first-child::before]:content-[attr(data-placeholder)]'
          )}
        />
      </div>
    </div>
  );
}
