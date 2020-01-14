FROM chromedp/headless-shell:latest
#FROM markadams/chromium-xvfb

RUN apt-get update && apt-get install -y \
    dumb-init

ENTRYPOINT ["dumb-init", "--"]
# or ENTRYPOINT ["tini", "--"]
CMD ["/usr/bin/google-chrome"]
