# koki/short usage

[short](github.com/koki/short) objects are used instead of native Kubernetes objects because they provide a cleaner interface and other conversion benefits down the road.  The objects are defined in the short/types directory.

## Help

Contact the koki/short developers for any issues related to the short objects.

## Reverting

If needed, reverting from koki/short objects to native kubernetes objects is fairly simple.  koki/short provides converters in the short/converter/converters directory.  Those can be used to translate the short objects back to native kubernetes object and evaluate where changes are needed to use native kubernetes objects.
