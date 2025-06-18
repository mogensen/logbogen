#!/bin/bash
set -e

CONVERT="docker run -u 1000 --rm -v $PWD/assets/images/:/imgs --entrypoint convert emiketic/image-processing"
COMPOSITE="docker run -u 1000 --rm -v $PWD/assets/images/:/imgs --entrypoint composite emiketic/image-processing"

declare -a ActivityArray=("boulder" "highrope" "ice" "other" "rock" "tree" "wall" "sail" "kayak" "canoe" "paddle-board")

arr=(assets/images/activities/*)

# iterate through array using a counter
for ((i=0; i<${#arr[@]}; i++)); do
    #do something to each element of array
    f=`basename ${arr[$i]} |cut -d"." -f1`
    ActivityArray+=($f)
done

# Create star PNG
$CONVERT -background none -size 1024x1024 /imgs/stars/star.svg  /imgs/stars/star.png

# Climbing icons star PNG and Gray
for activityType in "${ActivityArray[@]}"; do
  echo $activityType
  $CONVERT -background none -size 1024x1024 /imgs/activities/${activityType}.svg   /imgs/activities/${activityType}.png
  $CONVERT /imgs/activities/${activityType}.png -set colorspace Gray -average /imgs/activities/${activityType}_gray.png
done

# Rotate stars and make gray
for rotate in 0 72 144 216 288; do
  echo Star: $rotate
  $CONVERT /imgs/stars/star.png -background none -virtual-pixel background -distort ScaleRotateTranslate -$rotate /imgs/stars/star_$rotate.png
  $CONVERT /imgs/stars/star_$rotate.png -set colorspace Gray -average /imgs/stars/star_${rotate}_gray.png
done

# Create 5 levels for each climbingtype
for activityType in "${ActivityArray[@]}"; do
  echo Achievement for: $activityType

  $CONVERT -size 1024x1024  xc:none \
    /imgs/activities/${activityType}_gray.png  -composite \
    /imgs/stars/star_0_gray.png           -composite \
    /imgs/stars/star_72_gray.png          -composite \
    /imgs/stars/star_144_gray.png         -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${activityType}-0.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/activities/${activityType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72_gray.png          -composite \
    /imgs/stars/star_144_gray.png         -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${activityType}-1.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/activities/${activityType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144_gray.png         -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${activityType}-2.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/activities/${activityType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144.png              -composite \
    /imgs/stars/star_216_gray.png         -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${activityType}-3.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/activities/${activityType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144.png              -composite \
    /imgs/stars/star_216.png              -composite \
    /imgs/stars/star_288_gray.png         -composite \
    /imgs/achievements/${activityType}-4.png

  $CONVERT -size 1024x1024  xc:none \
    /imgs/activities/${activityType}.png       -composite \
    /imgs/stars/star_0.png                -composite \
    /imgs/stars/star_72.png               -composite \
    /imgs/stars/star_144.png              -composite \
    /imgs/stars/star_216.png              -composite \
    /imgs/stars/star_288.png              -composite \
    /imgs/achievements/${activityType}-5.png
done

rm ./assets/images/activities/*.png
rm ./assets/images/stars/*.png