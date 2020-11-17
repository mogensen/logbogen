#!/bin/bash

CONVERT="docker run -u 1000 --rm -v /home/mogensen/private/logbogen/assets/images/:/imgs --entrypoint convert emiketic/image-processing"
COMPOSITE="docker run -u 1000 --rm -v /home/mogensen/private/logbogen/assets/images/:/imgs --entrypoint composite emiketic/image-processing"

declare -a ClimbingArray=("Boulder" "HighRope" "Ice" "Other" "Rock" "Tree" "Wall")

# Create star PNG
$CONVERT -background none -size 1024x1024 /imgs/stars/star.svg  /imgs/stars/star.png

# Climbing icons star PNG and Gray
for climbType in "${ClimbingArray[@]}"; do
  echo $climbType
  $CONVERT -background none -size 1024x1024 /imgs/climbing/${climbType}.svg   /imgs/climbing/${climbType}.png
  $CONVERT /imgs/climbing/${climbType}.png -set colorspace Gray -average /imgs/climbing/${climbType}_gray.png
done

# Rotate stars and make gray
for rotate in 0 72 144 216 288; do
  echo Star: $rotate
  $CONVERT /imgs/stars/star.png -background none -virtual-pixel background -distort ScaleRotateTranslate -$rotate /imgs/stars/star_$rotate.png
  $CONVERT /imgs/stars/star_$rotate.png -set colorspace Gray -average /imgs/stars/star_${rotate}_gray.png
done

# Create 5 levels for each climbingtype
for climbType in "${ClimbingArray[@]}"; do
  echo Achievement for: $climbType

  $CONVERT -size 1024x1024  xc:none \
    /imgs/climbing/${climbType}_gray.png  -composite \
    /imgs/stars/star_0_gray.png           -composite \
    /imgs/stars/star_72_gray.png          -composite \
    /imgs/stars/star_144_gray.png         -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${climbType}-0.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/climbing/${climbType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72_gray.png          -composite \
    /imgs/stars/star_144_gray.png         -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${climbType}-1.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/climbing/${climbType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144_gray.png         -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${climbType}-2.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/climbing/${climbType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144.png              -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${climbType}-3.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/climbing/${climbType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144.png              -composite \
    /imgs/stars/star_216.png              -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${climbType}-4.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/climbing/${climbType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144.png              -composite \
    /imgs/stars/star_216.png              -composite \
    /imgs/stars/star_288.png              -composite \
    /imgs/achievements/${climbType}-5.png
done
