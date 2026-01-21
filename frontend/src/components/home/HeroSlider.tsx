'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { ChevronLeft, ChevronRight, BookOpen } from 'lucide-react';
import { useLocale } from 'next-intl';

export interface HeroSlideItem {
  id: string;
  slug: string;
  title: string;
  description?: string;
  coverUrl?: string;
}

export function HeroSlider({ slides, isLoading }: { slides: HeroSlideItem[]; isLoading?: boolean }) {
  const locale = useLocale();
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isAutoPlay, setIsAutoPlay] = useState(true);
  
  // Keep index in bounds when slides load/change
  useEffect(() => {
    if (slides.length === 0) return;
    setCurrentSlide((s) => Math.min(s, slides.length - 1));
  }, [slides.length]);

  // Autoplay
  useEffect(() => {
    if (!isAutoPlay) return;
    if (slides.length <= 1) return;
    
    const interval = setInterval(() => {
      setCurrentSlide((prev) => (prev + 1) % slides.length);
    }, 5000);

    return () => clearInterval(interval);
  }, [isAutoPlay, slides.length]);

  const goToSlide = (index: number) => {
    setCurrentSlide(index);
    setIsAutoPlay(false);
    setTimeout(() => setIsAutoPlay(true), 10000);
  };

  const nextSlide = () => goToSlide((currentSlide + 1) % slides.length);
  const prevSlide = () => goToSlide((currentSlide - 1 + slides.length) % slides.length);

  if (isLoading) {
    return (
      <div className="relative h-[400px] md:h-[500px] overflow-hidden bg-background-primary">
        <div className="container-custom h-full flex items-center">
          <div className="w-full grid grid-cols-1 md:grid-cols-2 gap-8 items-center animate-pulse">
            <div className="order-2 md:order-1">
              <div className="h-10 bg-background-hover rounded w-3/4 mb-4" />
              <div className="h-5 bg-background-hover rounded w-full mb-2" />
              <div className="h-5 bg-background-hover rounded w-5/6 mb-6" />
              <div className="h-12 bg-background-hover rounded w-40" />
            </div>
            <div className="order-1 md:order-2 flex justify-center md:justify-end">
              <div className="w-[180px] md:w-[220px] aspect-cover bg-background-hover rounded-card" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!slides || slides.length === 0) {
    return (
      <div className="relative h-[400px] md:h-[500px] overflow-hidden bg-background-primary">
        <div className="container-custom h-full flex items-center justify-center">
          <div className="text-center text-foreground-secondary">
            Пока нет новелл для показа
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="relative h-[400px] md:h-[500px] overflow-hidden bg-background-primary">
      {/* Slides */}
      {slides.map((slide, index) => (
        <div
          key={slide.id}
          className={`absolute inset-0 transition-opacity duration-500 ${
            index === currentSlide ? 'opacity-100 z-10' : 'opacity-0 z-0'
          }`}
        >
          {/* Background Image with Blur */}
          <div className="absolute inset-0">
            {slide.coverUrl ? (
              <Image
                src={slide.coverUrl}
                alt=""
                fill
                sizes="100vw"
                className="object-cover blur-lg scale-110"
                priority={index === 0}
              />
            ) : null}
            <div className="absolute inset-0 bg-gradient-to-r from-background-primary via-background-primary/90 to-transparent" />
          </div>

          {/* Content */}
          <div className="container-custom relative z-10 h-full flex items-center">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 items-center">
              {/* Text Content */}
              <div className="order-2 md:order-1">
                <h2 className="text-3xl md:text-4xl lg:text-5xl font-heading font-bold mb-4 line-clamp-2">
                  {slide.title}
                </h2>
                <p className="text-foreground-secondary text-lg mb-6 line-clamp-3">
                  {slide.description || ''}
                </p>
                <div className="flex items-center gap-4">
                  <Link
                    href={`/${locale}/novel/${slide.slug}`}
                    className="btn-primary text-base px-6 py-3"
                  >
                    <BookOpen className="w-5 h-5 mr-2" />
                    Читать
                  </Link>
                </div>
              </div>

              {/* Cover Image */}
              <div className="order-1 md:order-2 flex justify-center md:justify-end">
                <Link
                  href={`/${locale}/novel/${slide.slug}`}
                  className="relative w-[180px] md:w-[220px] aspect-cover rounded-card overflow-hidden shadow-card-hover hover-lift"
                >
                  {slide.coverUrl ? (
                    <Image
                      src={slide.coverUrl}
                      alt={slide.title}
                      fill
                      sizes="(max-width: 768px) 180px, 220px"
                      className="object-cover"
                      priority={index === 0}
                    />
                  ) : (
                    <div className="w-full h-full bg-background-tertiary flex items-center justify-center">
                      <BookOpen className="w-12 h-12 text-foreground-muted" />
                    </div>
                  )}
                </Link>
              </div>
            </div>
          </div>
        </div>
      ))}

      {/* Navigation Arrows */}
      {slides.length > 1 && (
      <button
        onClick={prevSlide}
        className="absolute left-4 top-1/2 -translate-y-1/2 z-20 btn-ghost p-2 bg-background-primary/50 backdrop-blur-sm rounded-full"
        aria-label="Previous slide"
      >
        <ChevronLeft className="w-6 h-6" />
      </button>
      )}
      {slides.length > 1 && (
      <button
        onClick={nextSlide}
        className="absolute right-4 top-1/2 -translate-y-1/2 z-20 btn-ghost p-2 bg-background-primary/50 backdrop-blur-sm rounded-full"
        aria-label="Next slide"
      >
        <ChevronRight className="w-6 h-6" />
      </button>
      )}

      {/* Dots Indicator */}
      {slides.length > 1 && (
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 z-20 flex gap-2">
          {slides.map((_, index) => (
            <button
              key={index}
              onClick={() => goToSlide(index)}
              className={`w-2 h-2 rounded-full transition-all ${
                index === currentSlide
                  ? 'w-8 bg-accent-primary'
                  : 'bg-foreground-muted/50 hover:bg-foreground-muted'
              }`}
              aria-label={`Go to slide ${index + 1}`}
            />
          ))}
        </div>
      )}
    </div>
  );
}
