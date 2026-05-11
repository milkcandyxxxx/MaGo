// MaGo (マゴ) - Main JavaScript
// 适量动画 + 交互功能

document.addEventListener('DOMContentLoaded', function() {
    initLoading();
    initScrollAnimations();
    initHeaderScroll();
    initMobileMenu();
    initImagePreview();
    initBackToTop();
    initMouseGlow();
    initTypewriter();
    initStaggerAnimations();
});

// ============================================
// Loading Animation
// ============================================
function initLoading() {
    const loading = document.getElementById('loading');
    if (!loading) return;

    window.addEventListener('load', function() {
        setTimeout(function() {
            loading.classList.add('hidden');
            setTimeout(function() {
                loading.style.display = 'none';
            }, 500);
        }, 300);
    });
}

// ============================================
// Scroll Animations
// ============================================
function initScrollAnimations() {
    const elements = document.querySelectorAll('.animate-on-scroll');
    if (elements.length === 0) return;

    const observer = new IntersectionObserver(function(entries) {
        entries.forEach(function(entry) {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
            }
        });
    }, {
        threshold: 0.1,
        rootMargin: '0px 0px -30px 0px'
    });

    elements.forEach(function(el) {
        observer.observe(el);
    });
}

// ============================================
// Header Scroll Effect
// ============================================
function initHeaderScroll() {
    const header = document.getElementById('header');
    if (!header) return;

    let ticking = false;

    window.addEventListener('scroll', function() {
        if (!ticking) {
            requestAnimationFrame(function() {
                if (window.scrollY > 50) {
                    header.classList.add('scrolled');
                } else {
                    header.classList.remove('scrolled');
                }
                ticking = false;
            });
            ticking = true;
        }
    });
}

// ============================================
// Mobile Menu
// ============================================
function initMobileMenu() {
    const btn = document.querySelector('.mobile-menu-btn');
    const nav = document.querySelector('.nav-right');
    if (!btn || !nav) return;

    btn.addEventListener('click', function() {
        nav.classList.toggle('active');
    });

    // 点击菜单项后关闭菜单
    nav.querySelectorAll('.nav-link').forEach(function(link) {
        link.addEventListener('click', function() {
            nav.classList.remove('active');
        });
    });
}

// ============================================
// Image Preview
// ============================================
function initImagePreview() {
    document.querySelectorAll('.article-content img').forEach(function(img) {
        img.style.cursor = 'pointer';

        img.addEventListener('click', function() {
            const overlay = document.createElement('div');
            overlay.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.9);display:flex;align-items:center;justify-content:center;z-index:10000;cursor:pointer;animation:fadeIn 0.3s ease;';

            const previewImg = document.createElement('img');
            previewImg.src = this.src;
            previewImg.style.cssText = 'max-width:90%;max-height:90%;border-radius:10px;animation:scaleIn 0.3s ease;';

            overlay.appendChild(previewImg);
            document.body.appendChild(overlay);

            overlay.addEventListener('click', function() {
                overlay.style.opacity = '0';
                overlay.style.transition = 'opacity 0.3s';
                setTimeout(function() {
                    overlay.remove();
                }, 300);
            });
        });
    });
}

// ============================================
// Back to Top
// ============================================
function initBackToTop() {
    const btn = document.createElement('button');
    btn.innerHTML = '&#8593;';
    btn.style.cssText = 'position:fixed;bottom:30px;right:30px;background:linear-gradient(135deg,#9abbf7,#ffa2c4);border:none;border-radius:50%;color:#fff;cursor:pointer;height:45px;width:45px;font-size:20px;display:none;align-items:center;justify-content:center;box-shadow:0 4px 12px rgba(154,187,247,0.4);transition:all 0.25s;z-index:1000;';
    btn.className = 'back-to-top';

    document.body.appendChild(btn);

    window.addEventListener('scroll', function() {
        btn.style.display = window.scrollY > 300 ? 'flex' : 'none';
    });

    btn.addEventListener('click', function() {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });

    btn.addEventListener('mouseenter', function() {
        this.style.transform = 'translateY(-3px)';
        this.style.boxShadow = '0 6px 16px rgba(154,187,247,0.6)';
    });

    btn.addEventListener('mouseleave', function() {
        this.style.transform = 'translateY(0)';
        this.style.boxShadow = '0 4px 12px rgba(154,187,247,0.4)';
    });
}

// ============================================
// Mouse Glow Effect
// ============================================
function initMouseGlow() {
    document.querySelectorAll('.mouse-glow').forEach(function(el) {
        el.addEventListener('mousemove', function(e) {
            const rect = this.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;
            this.style.setProperty('--mouse-x', x + 'px');
            this.style.setProperty('--mouse-y', y + 'px');
        });
    });
}

// ============================================
// Typewriter Effect
// ============================================
function initTypewriter() {
    document.querySelectorAll('.typewriter-auto').forEach(function(el) {
        const text = el.textContent;
        el.textContent = '';
        el.style.visibility = 'visible';

        let i = 0;
        function type() {
            if (i < text.length) {
                el.textContent += text.charAt(i);
                i++;
                setTimeout(type, 100);
            }
        }

        // 当元素进入视口时开始打字
        const observer = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    type();
                    observer.unobserve(el);
                }
            });
        });

        observer.observe(el);
    });
}

// ============================================
// Stagger Animations
// ============================================
function initStaggerAnimations() {
    document.querySelectorAll('.stagger-container').forEach(function(container) {
        const children = container.children;

        const observer = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    Array.from(children).forEach(function(child, index) {
                        child.style.animationDelay = (index * 0.1) + 's';
                        child.classList.add('anim-fade-in-up', 'anim-fill-both');
                    });
                    observer.unobserve(container);
                }
            });
        }, { threshold: 0.1 });

        observer.observe(container);
    });
}

// ============================================
// Counter Animation
// ============================================
function animateCounter(element, target, duration) {
    let start = 0;
    const increment = target / (duration / 16);

    function update() {
        start += increment;
        if (start < target) {
            element.textContent = Math.floor(start);
            requestAnimationFrame(update);
        } else {
            element.textContent = target;
        }
    }

    update();
}

// 初始化计数器动画
document.querySelectorAll('.counter').forEach(function(el) {
    const target = parseInt(el.getAttribute('data-target'));

    const observer = new IntersectionObserver(function(entries) {
        entries.forEach(function(entry) {
            if (entry.isIntersecting) {
                animateCounter(el, target, 1000);
                observer.unobserve(el);
            }
        });
    });

    observer.observe(el);
});
