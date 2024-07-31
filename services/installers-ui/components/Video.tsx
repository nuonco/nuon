import React, { type FC } from "react";

interface IVideo {
  className?: string;
  src: string;
}

export const Video: FC<IVideo> = ({ className = "", src }) => {
  return (
    <div className="relative pb-[56.25%] h-0 overflow-hidden max-w-full rounded-video-container-radius">
      <iframe
        className={`absolute top-0 left-0 w-full h-full ${className}`}
        src={src}
        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
        allowFullScreen
      ></iframe>
    </div>
  );
};
