'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { ChevronLeft, ChevronRight, BookOpen } from 'lucide-react';
import { useLocale } from 'next-intl';

interface SlideItem {
  id: string;
  slug: string;
  title: string;
  description: string;
  coverUrl: string;
  latestChapter?: number;
}

export function HeroSlider() {
  const locale = useLocale();
  const [currentSlide, setCurrentSlide] = useState(0);
  const [isAutoPlay, setIsAutoPlay] = useState(true);

  // Mock data - TODO: fetch from API
  const slides: SlideItem[] = [
    {
      id: '1',
      slug: 'the-beginning-after-the-end',
      title: 'Начало после конца',
      description: 'Король Грей обладает непревзойденной силой и престижем. Однако одиночество следует за теми, кто обладает большой властью...',
      coverUrl: '/placeholder-hero-1.svg',
      latestChapter: 450,
    },
    {
      id: '2',
      slug: 'solo-leveling',
      title: 'Поднятие уровня в одиночку',
      description: 'В мире, где пробудились охотники, обладающие магическими способностями, Сон Джин-ву — слабейший среди них...',
      coverUrl: '/placeholder-hero-2.svg',
      latestChapter: 270,
    },
    {
      id: '3',
      slug: 'omniscient-readers-viewpoint',
      title: 'Точка зрения всеведущего читателя',
      description: 'Единственный читатель, закончивший роман "Три способа выжить в разрушенном мире", Ким Доккджа...',
      coverUrl: '/placeholder-hero-3.svg',
      latestChapter: 551,
    },
  ];

  // Autoplay
  useEffect(() => {
    if (!isAutoPlay) return;
    
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
            <Image
              src={slide.coverUrl}
              alt=""
              fill
              className="object-cover blur-lg scale-110"
              priority={index === 0}
            />
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
                  {slide.description}
                </p>
                <div className="flex items-center gap-4">
                  <Link
                    href={`/${locale}/novel/${slide.slug}`}
                    className="btn-primary text-base px-6 py-3"
                  >
                    <BookOpen className="w-5 h-5 mr-2" />
                    Читать
                  </Link>
                  {slide.latestChapter && (
                    <span className="text-foreground-secondary">
                      Глава {slide.latestChapter}
                    </span>
                  )}
                </div>
              </div>

              {/* Cover Image */}
              <div className="order-1 md:order-2 flex justify-center md:justify-end">
                <Link
                  href={`/${locale}/novel/${slide.slug}`}
                  className="relative w-[180px] md:w-[220px] aspect-cover rounded-card overflow-hidden shadow-card-hover hover-lift"
                >
                  <Image
                    src={slide.coverUrl}
                    alt={slide.title}
                    fill
                    className="object-cover"
                    priority={index === 0}
                  />
                </Link>
              </div>
            </div>
          </div>
        </div>
      ))}

      {/* Navigation Arrows */}
      <button
        onClick={prevSlide}
        className="absolute left-4 top-1/2 -translate-y-1/2 z-20 btn-ghost p-2 bg-background-primary/50 backdrop-blur-sm rounded-full"
        aria-label="Previous slide"
      >
        <ChevronLeft className="w-6 h-6" />
      </button>
      <button
        onClick={nextSlide}
        className="absolute right-4 top-1/2 -translate-y-1/2 z-20 btn-ghost p-2 bg-background-primary/50 backdrop-blur-sm rounded-full"
        aria-label="Next slide"
      >
        <ChevronRight className="w-6 h-6" />
      </button>

      {/* Dots Indicator */}
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
    </div>
  );
}
