import { useEffect } from "react";
import { X } from "lucide-react";
import { TransformWrapper, TransformComponent } from "react-zoom-pan-pinch";

interface Props {
  src: string;
  alt: string;
  caption?: string | null;
  onClose: () => void;
}

export function Lightbox({ src, alt, caption, onClose }: Props) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = prevOverflow;
    };
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 bg-black/95 flex items-center justify-center touch-none select-none"
      onClick={onClose}
    >
      <button
        type="button"
        onClick={onClose}
        aria-label="閉じる"
        className="absolute top-4 right-4 z-10 p-2 rounded-full bg-zinc-900/80 text-zinc-200 hover:bg-zinc-800 active:bg-zinc-700 transition-colors"
      >
        <X size={22} />
      </button>

      {/* 内部は伝播を止めて、画像操作中に背景タップで閉じないようにする */}
      <div className="w-full h-full" onClick={(e) => e.stopPropagation()}>
        <TransformWrapper
          initialScale={1}
          minScale={1}
          maxScale={6}
          doubleClick={{ mode: "toggle", step: 2 }}
          pinch={{ step: 5 }}
          wheel={{ step: 0.2 }}
          limitToBounds
          centerOnInit
        >
          <TransformComponent
            wrapperStyle={{ width: "100%", height: "100%" }}
            contentStyle={{ width: "100%", height: "100%" }}
          >
            <img
              src={src}
              alt={alt}
              className="w-full h-full object-contain"
              draggable={false}
            />
          </TransformComponent>
        </TransformWrapper>
      </div>

      {caption && (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-zinc-900/80 backdrop-blur-md text-yellow-400 text-xs font-bold tracking-wider px-3 py-1.5 rounded-md border border-yellow-500/20">
          {caption}
        </div>
      )}
    </div>
  );
}
