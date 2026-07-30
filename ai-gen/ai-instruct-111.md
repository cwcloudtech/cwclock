# AI instruction 111

## Mobile CICD

I want you to fix all the following errors (coming from the CICD Docker build):

```shell
#31 280.5 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:1187:23: warning: the variable "DebuggerInternal" was not declared in function "__shouldPauseOnThrow"
#31 280.5         return typeof DebuggerInternal !== 'undefined' && DebuggerInternal.sh...
#31 280.5                       ^~~~~~~~~~~~~~~~
#31 280.5 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:5529:7: warning: the variable "setTimeout" was not declared in function "logCapturedError"
#31 280.5       setTimeout(function () {
#31 280.5       ^~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:3763:31: warning: the variable "nativeFabricUIManager" was not declared in anonymous function " 134#"
#31 280.6   var _nativeFabricUIManage = nativeFabricUIManager,
#31 280.6                               ^~~~~~~~~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:3791:21: warning: the variable "clearTimeout" was not declared in anonymous function " 134#"
#31 280.6     cancelTimeout = clearTimeout;
#31 280.6                     ^~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:3804:51: warning: the variable "RN$enableMicrotasksInReact" was not declared in anonymous function " 134#"
#31 280.6 ... "undefined" !== typeof RN$enableMicrotasksInReact && !!RN$enableMicrotask...
#31 280.6                            ^~~~~~~~~~~~~~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:3805:47: warning: the variable "queueMicrotask" was not declared in anonymous function " 134#"
#31 280.6 ...otask = "function" === typeof queueMicrotask ? queueMicrotask : scheduleTi...
#31 280.6                                  ^~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:8226:30: warning: the variable "__REACT_DEVTOOLS_GLOBAL_HOOK__" was not declared in anonymous function " 134#"
#31 280.6   if ("undefined" !== typeof __REACT_DEVTOOLS_GLOBAL_HOOK__) {
#31 280.6                              ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:10267:5: warning: the variable "setImmediate" was not declared in function "handleResolved"
#31 280.6     setImmediate(function () {
#31 280.6     ^~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:13982:12: warning: the variable "fetch" was not declared in anonymous function " 388#"
#31 280.6     fetch: fetch,
#31 280.6            ^~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:13983:14: warning: the variable "Headers" was not declared in anonymous function " 388#"
#31 280.6     Headers: Headers,
#31 280.6              ^~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:13984:14: warning: the variable "Request" was not declared in anonymous function " 388#"
#31 280.6     Request: Request,
#31 280.6              ^~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:13985:15: warning: the variable "Response" was not declared in anonymous function " 388#"
#31 280.6     Response: Response
#31 280.6               ^~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:14142:24: warning: the variable "FileReader" was not declared in function "readBlobAsArrayBuffer"
#31 280.6       var reader = new FileReader();
#31 280.6                        ^~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:14193:36: warning: the variable "Blob" was not declared in anonymous function " 399#"
#31 280.6         } else if (support.blob && Blob.prototype.isPrototypeOf(body)) {
#31 280.6                                    ^~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:14195:40: warning: the variable "FormData" was not declared in anonymous function " 399#"
#31 280.6         } else if (support.formData && FormData.prototype.isPrototypeOf(body)) {
#31 280.6                                        ^~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:14197:44: warning: the variable "URLSearchParams" was not declared in anonymous function " 399#"
#31 280.6 ...e if (support.searchParams && URLSearchParams.prototype.isPrototypeOf(body...
#31 280.6                                  ^~~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:14316:26: warning: the variable "AbortController" was not declared in anonymous function " 405#"
#31 280.6           var ctrl = new AbortController();
#31 280.6                          ^~~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:14450:23: warning: the variable "XMLHttpRequest" was not declared in anonymous function " 409#"
#31 280.6         var xhr = new XMLHttpRequest();
#31 280.6                       ^~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:13995:71: warning: the variable "self" was not declared in anonymous function " 391#"
#31 280.6 ...undefined' && globalThis || typeof self !== 'undefined' && self ||
#31 280.6                                       ^~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:22181:26: warning: the variable "navigator" was not declared in anonymous function " 701#"
#31 280.6   "undefined" !== typeof navigator && undefined !== navigator.scheduling && u...
#31 280.6                          ^~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:22291:37: warning: the variable "MessageChannel" was not declared in anonymous function " 701#"
#31 280.6   };else if ("undefined" !== typeof MessageChannel) {
#31 280.6                                     ^~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:22306:34: warning: the variable "nativeRuntimeScheduler" was not declared in anonymous function " 701#"
#31 280.6 ... = "undefined" !== typeof nativeRuntimeScheduler ? nativeRuntimeScheduler....
#31 280.6                              ^~~~~~~~~~~~~~~~~~~~~~
#31 280.6 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:30391:34: warning: the variable "requestAnimationFrame" was not declared in function "start 9#"
#31 280.6 ...    this._animationFrame = requestAnimationFrame(this.onUpdate.bind(this));
#31 280.6                               ^~~~~~~~~~~~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:49537:49: warning: the variable "btoa" was not declared in function "resolveConfig"
#31 280.7 ...rs.set('Authorization', 'Basic ' + btoa(username + ':' + (password ? encod...
#31 280.7                                       ^~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:48711:19: warning: the variable "WorkerGlobalScope" was not declared in anonymous function " 1555#"
#31 280.7     return typeof WorkerGlobalScope !== 'undefined' &&
#31 280.7                   ^~~~~~~~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:49153:17: warning: the variable "URL" was not declared in anonymous function " 1563#"
#31 280.7       url = new URL(url, platform.origin);
#31 280.7                 ^~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:49880:16: warning: the variable "ReadableStream" was not declared in function "trackStream"
#31 280.7     return new ReadableStream({
#31 280.7                ^~~~~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:79408:9: warning: the variable "REACT_NAVIGATION_DEVTOOLS" was not declared in anonymous function " 3041#"
#31 280.7         REACT_NAVIGATION_DEVTOOLS.set(refContainer.current, {
#31 280.7         ^~~~~~~~~~~~~~~~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:82569:73: warning: the variable "HTMLElement" was not declared in anonymous function " 3189#"
#31 280.7 ...Platform.OS === 'web' && typeof HTMLElement !== 'undefined' && node instan...
#31 280.7                                    ^~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:83213:26: warning: the variable "ResizeObserver" was not declared in anonymous function " 3213#"
#31 280.7       var observer = new ResizeObserver(function (entries) {
#31 280.7                          ^~~~~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:84293:16: warning: the variable "requestIdleCallback" was not declared in anonymous function " 3255#"
#31 280.7       var id = requestIdleCallback(function () {
#31 280.7                ^~~~~~~~~~~~~~~~~~~
#31 280.7 /app/android/app/build/generated/assets/createBundleReleaseJsAndAssets/index.android.bundle:84297:16: warning: the variable "cancelIdleCallback" was not declared in anonymous function " 3257#"
#31 280.7         return cancelIdleCallback(id);
#31 280.7                ^~~~~~~~~~~~~~~~~~
#31 283.3 
#31 283.3 > Task :react-native-camera-kit:verifyReleaseResources
#31 283.3 > Task :react-native-camera-kit:assembleRelease
#31 283.4 > Task :react-native-community_datetimepicker:javaPreCompileRelease
#31 283.5 > Task :react-native-community_datetimepicker:mergeReleaseShaders
#31 283.5 > Task :react-native-community_datetimepicker:compileReleaseShaders NO-SOURCE
#31 283.5 > Task :react-native-community_datetimepicker:generateReleaseAssets UP-TO-DATE
#31 283.6 > Task :react-native-community_datetimepicker:packageReleaseAssets
#31 283.7 > Task :react-native-community_datetimepicker:prepareLintJarForPublish
#31 283.8 > Task :react-native-community_datetimepicker:prepareReleaseArtProfile
#31 283.9 > Task :react-native-community_datetimepicker:processReleaseManifest
#31 284.0 > Task :react-native-community_datetimepicker:writeReleaseAarMetadata
#31 287.3 > Task :react-native-community_datetimepicker:mapReleaseSourceSetPaths
#31 287.3 > Task :react-native-community_datetimepicker:compileReleaseKotlin
#31 287.4 > Task :react-native-gesture-handler:javaPreCompileRelease
#31 287.5 > Task :react-native-gesture-handler:mergeReleaseShaders
#31 287.5 > Task :react-native-gesture-handler:compileReleaseShaders NO-SOURCE
#31 287.5 > Task :react-native-gesture-handler:generateReleaseAssets UP-TO-DATE
#31 287.7 > Task :react-native-gesture-handler:packageReleaseAssets
#31 287.8 > Task :react-native-gesture-handler:prepareLintJarForPublish
#31 287.8 > Task :react-native-gesture-handler:prepareReleaseArtProfile
#31 288.0 > Task :react-native-gesture-handler:processReleaseManifest
#31 288.0 > Task :react-native-gesture-handler:writeReleaseAarMetadata
#31 288.0 > Task :app:generateReleaseResValues
#31 288.0 > Task :react-native-gesture-handler:mapReleaseSourceSetPaths
#31 288.1 > Task :app:mapReleaseSourceSetPaths
#31 288.2 > Task :app:generateReleaseResources
#31 289.0 
#31 289.0 > Task :app:mergeReleaseResources FAILED
#31 289.0 ERROR: /app/android/app/src/main/res/values/colors.xml:2:33: Resource and asset merger: The string "--" is not permitted within comments.
#31 289.0     org.xml.sax.SAXParseException; lineNumber: 2; columnNumber: 33; The string "--" is not permitted within comments.
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.util.ErrorHandlerWrapper.createSAXParseException(ErrorHandlerWrapper.java:204)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.util.ErrorHandlerWrapper.fatalError(ErrorHandlerWrapper.java:178)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLErrorReporter.reportError(XMLErrorReporter.java:400)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLErrorReporter.reportError(XMLErrorReporter.java:327)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLScanner.reportFatalError(XMLScanner.java:1465)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLScanner.scanComment(XMLScanner.java:818)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLDocumentFragmentScannerImpl.scanComment(XMLDocumentFragmentScannerImpl.java:1079)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLDocumentFragmentScannerImpl$FragmentContentDriver.next(XMLDocumentFragmentScannerImpl.java:2914)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLDocumentScannerImpl.next(XMLDocumentScannerImpl.java:605)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLNSDocumentScannerImpl.next(XMLNSDocumentScannerImpl.java:112)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.impl.XMLDocumentFragmentScannerImpl.scanDocument(XMLDocumentFragmentScannerImpl.java:542)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.parsers.XML11Configuration.parse(XML11Configuration.java:889)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.parsers.XML11Configuration.parse(XML11Configuration.java:825)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.parsers.XMLParser.parse(XMLParser.java:141)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.parsers.AbstractSAXParser.parse(AbstractSAXParser.java:1224)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.jaxp.SAXParserImpl$JAXPSAXParser.parse(SAXParserImpl.java:637)
#31 289.0     	at java.xml/com.sun.org.apache.xerces.internal.jaxp.SAXParserImpl.parse(SAXParserImpl.java:326)
#31 289.0     	at com.android.utils.PositionXmlParser.parseInternal(PositionXmlParser.java:284)
#31 289.0     	at com.android.utils.PositionXmlParser.parseInternal(PositionXmlParser.java:233)
#31 289.0     	at com.android.utils.PositionXmlParser.parse(PositionXmlParser.java:179)
#31 289.0     	at com.android.utils.PositionXmlParser.parse(PositionXmlParser.java:104)
#31 289.0     	at com.android.utils.PositionXmlParser.parse(PositionXmlParser.java:143)
#31 289.0     	at com.android.ide.common.resources.ValueResourceParser2.parseDocument(ValueResourceParser2.java:216)
#31 289.0     	at com.android.ide.common.resources.ValueResourceParser2.parseFile(ValueResourceParser2.java:92)
#31 289.0     	at com.android.ide.common.resources.ResourceSet.createResourceFile(ResourceSet.java:560)
#31 289.0     	at com.android.ide.common.resources.ResourceSet.parseFolder(ResourceSet.java:487)
#31 289.0     	at com.android.ide.common.resources.ResourceSet.readSourceFolder(ResourceSet.java:284)
#31 289.0     	at com.android.ide.common.resources.DataSet.loadFromFiles(DataSet.java:262)
#31 289.0     	at com.android.ide.common.resources.DataSet.loadFromFiles(DataSet.java:243)
#31 289.0     	at com.android.build.gradle.tasks.MergeResources$doFullTaskAction$1$1$1.invoke(MergeResources.kt:243)
#31 289.0     	at com.android.build.gradle.internal.tasks.Blocks.recordSpan(Blocks.java:51)
#31 289.0     	at com.android.build.gradle.tasks.MergeResources.doFullTaskAction(MergeResources.kt:237)
#31 289.0     	at com.android.build.gradle.tasks.MergeResources.doTaskAction(MergeResources.kt:322)
#31 289.0     	at com.android.build.gradle.internal.tasks.NewIncrementalTask$taskAction$$inlined$recordTaskAction$1.invoke(BaseTask.kt:70)
#31 289.0     	at com.android.build.gradle.internal.tasks.Blocks.recordSpan(Blocks.java:51)
#31 289.0     	at com.android.build.gradle.internal.tasks.NewIncrementalTask.taskAction(NewIncrementalTask.kt:46)
#31 289.0     	at jdk.internal.reflect.GeneratedMethodAccessor317.invoke(Unknown Source)
#31 289.0     	at java.base/jdk.internal.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
#31 289.0     	at java.base/java.lang.reflect.Method.invoke(Method.java:569)
#31 289.0     	at org.gradle.internal.reflect.JavaMethod.invoke(JavaMethod.java:125)
#31 289.0     	at org.gradle.api.internal.project.taskfactory.IncrementalTaskAction.doExecute(IncrementalTaskAction.java:45)
#31 289.0     	at org.gradle.api.internal.project.taskfactory.StandardTaskAction.execute(StandardTaskAction.java:51)
#31 289.0     	at org.gradle.api.internal.project.taskfactory.IncrementalTaskAction.execute(IncrementalTaskAction.java:26)
#31 289.0     	at org.gradle.api.internal.project.taskfactory.StandardTaskAction.execute(StandardTaskAction.java:29)
#31 289.0     	at org.gradle.api.internal.tasks.execution.TaskExecution$3.run(TaskExecution.java:244)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$1.execute(DefaultBuildOperationRunner.java:29)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$1.execute(DefaultBuildOperationRunner.java:26)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:66)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:166)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.run(DefaultBuildOperationRunner.java:47)
#31 289.0     	at org.gradle.api.internal.tasks.execution.TaskExecution.executeAction(TaskExecution.java:229)
#31 289.0     	at org.gradle.api.internal.tasks.execution.TaskExecution.executeActions(TaskExecution.java:212)
#31 289.0     	at org.gradle.api.internal.tasks.execution.TaskExecution.executeWithPreviousOutputFiles(TaskExecution.java:195)
#31 289.0     	at org.gradle.api.internal.tasks.execution.TaskExecution.execute(TaskExecution.java:162)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteStep.executeInternal(ExecuteStep.java:105)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteStep.access$000(ExecuteStep.java:44)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteStep$1.call(ExecuteStep.java:59)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteStep$1.call(ExecuteStep.java:56)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$CallableBuildOperationWorker.execute(DefaultBuildOperationRunner.java:209)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$CallableBuildOperationWorker.execute(DefaultBuildOperationRunner.java:204)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:66)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:166)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.call(DefaultBuildOperationRunner.java:53)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteStep.execute(ExecuteStep.java:56)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteStep.execute(ExecuteStep.java:44)
#31 289.0     	at org.gradle.internal.execution.steps.CancelExecutionStep.execute(CancelExecutionStep.java:42)
#31 289.0     	at org.gradle.internal.execution.steps.TimeoutStep.executeWithoutTimeout(TimeoutStep.java:75)
#31 289.0     	at org.gradle.internal.execution.steps.TimeoutStep.execute(TimeoutStep.java:55)
#31 289.0     	at org.gradle.internal.execution.steps.PreCreateOutputParentsStep.execute(PreCreateOutputParentsStep.java:50)
#31 289.0     	at org.gradle.internal.execution.steps.PreCreateOutputParentsStep.execute(PreCreateOutputParentsStep.java:28)
#31 289.0     	at org.gradle.internal.execution.steps.RemovePreviousOutputsStep.execute(RemovePreviousOutputsStep.java:67)
#31 289.0     	at org.gradle.internal.execution.steps.RemovePreviousOutputsStep.execute(RemovePreviousOutputsStep.java:37)
#31 289.0     	at org.gradle.internal.execution.steps.BroadcastChangingOutputsStep.execute(BroadcastChangingOutputsStep.java:61)
#31 289.0     	at org.gradle.internal.execution.steps.BroadcastChangingOutputsStep.execute(BroadcastChangingOutputsStep.java:26)
#31 289.0     	at org.gradle.internal.execution.steps.CaptureOutputsAfterExecutionStep.execute(CaptureOutputsAfterExecutionStep.java:69)
#31 289.0     	at org.gradle.internal.execution.steps.CaptureOutputsAfterExecutionStep.execute(CaptureOutputsAfterExecutionStep.java:46)
#31 289.0     	at org.gradle.internal.execution.steps.ResolveInputChangesStep.execute(ResolveInputChangesStep.java:40)
#31 289.0     	at org.gradle.internal.execution.steps.ResolveInputChangesStep.execute(ResolveInputChangesStep.java:29)
#31 289.0     	at org.gradle.internal.execution.steps.BuildCacheStep.executeWithoutCache(BuildCacheStep.java:189)
#31 289.0     	at org.gradle.internal.execution.steps.BuildCacheStep.lambda$execute$1(BuildCacheStep.java:75)
#31 289.0     	at org.gradle.internal.Either$Right.fold(Either.java:175)
#31 289.0     	at org.gradle.internal.execution.caching.CachingState.fold(CachingState.java:62)
#31 289.0     	at org.gradle.internal.execution.steps.BuildCacheStep.execute(BuildCacheStep.java:73)
#31 289.0     	at org.gradle.internal.execution.steps.BuildCacheStep.execute(BuildCacheStep.java:48)
#31 289.0     	at org.gradle.internal.execution.steps.StoreExecutionStateStep.execute(StoreExecutionStateStep.java:46)
#31 289.0     	at org.gradle.internal.execution.steps.StoreExecutionStateStep.execute(StoreExecutionStateStep.java:35)
#31 289.0     	at org.gradle.internal.execution.steps.SkipUpToDateStep.executeBecause(SkipUpToDateStep.java:75)
#31 289.0     	at org.gradle.internal.execution.steps.SkipUpToDateStep.lambda$execute$2(SkipUpToDateStep.java:53)
#31 289.0     	at java.base/java.util.Optional.orElseGet(Optional.java:364)
#31 289.0     	at org.gradle.internal.execution.steps.SkipUpToDateStep.execute(SkipUpToDateStep.java:53)
#31 289.0     	at org.gradle.internal.execution.steps.SkipUpToDateStep.execute(SkipUpToDateStep.java:35)
#31 289.0     	at org.gradle.internal.execution.steps.legacy.MarkSnapshottingInputsFinishedStep.execute(MarkSnapshottingInputsFinishedStep.java:37)
#31 289.0     	at org.gradle.internal.execution.steps.legacy.MarkSnapshottingInputsFinishedStep.execute(MarkSnapshottingInputsFinishedStep.java:27)
#31 289.0     	at org.gradle.internal.execution.steps.ResolveIncrementalCachingStateStep.executeDelegate(ResolveIncrementalCachingStateStep.java:49)
#31 289.0     	at org.gradle.internal.execution.steps.ResolveIncrementalCachingStateStep.executeDelegate(ResolveIncrementalCachingStateStep.java:27)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractResolveCachingStateStep.execute(AbstractResolveCachingStateStep.java:71)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractResolveCachingStateStep.execute(AbstractResolveCachingStateStep.java:39)
#31 289.0     	at org.gradle.internal.execution.steps.ResolveChangesStep.execute(ResolveChangesStep.java:65)
#31 289.0     	at org.gradle.internal.execution.steps.ResolveChangesStep.execute(ResolveChangesStep.java:36)
#31 289.0     	at org.gradle.internal.execution.steps.ValidateStep.execute(ValidateStep.java:107)
#31 289.0     	at org.gradle.internal.execution.steps.ValidateStep.execute(ValidateStep.java:56)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractCaptureStateBeforeExecutionStep.execute(AbstractCaptureStateBeforeExecutionStep.java:64)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractCaptureStateBeforeExecutionStep.execute(AbstractCaptureStateBeforeExecutionStep.java:43)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractSkipEmptyWorkStep.executeWithNonEmptySources(AbstractSkipEmptyWorkStep.java:125)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractSkipEmptyWorkStep.execute(AbstractSkipEmptyWorkStep.java:56)
#31 289.0     	at org.gradle.internal.execution.steps.AbstractSkipEmptyWorkStep.execute(AbstractSkipEmptyWorkStep.java:36)
#31 289.0     	at org.gradle.internal.execution.steps.legacy.MarkSnapshottingInputsStartedStep.execute(MarkSnapshottingInputsStartedStep.java:38)
#31 289.0     	at org.gradle.internal.execution.steps.LoadPreviousExecutionStateStep.execute(LoadPreviousExecutionStateStep.java:36)
#31 289.0     	at org.gradle.internal.execution.steps.LoadPreviousExecutionStateStep.execute(LoadPreviousExecutionStateStep.java:23)
#31 289.0     	at org.gradle.internal.execution.steps.HandleStaleOutputsStep.execute(HandleStaleOutputsStep.java:75)
#31 289.0     	at org.gradle.internal.execution.steps.HandleStaleOutputsStep.execute(HandleStaleOutputsStep.java:41)
#31 289.0     	at org.gradle.internal.execution.steps.AssignMutableWorkspaceStep.lambda$execute$0(AssignMutableWorkspaceStep.java:35)
#31 289.0     	at org.gradle.api.internal.tasks.execution.TaskExecution$4.withWorkspace(TaskExecution.java:289)
#31 289.0     	at org.gradle.internal.execution.steps.AssignMutableWorkspaceStep.execute(AssignMutableWorkspaceStep.java:31)
#31 289.0     	at org.gradle.internal.execution.steps.AssignMutableWorkspaceStep.execute(AssignMutableWorkspaceStep.java:22)
#31 289.0     	at org.gradle.internal.execution.steps.ChoosePipelineStep.execute(ChoosePipelineStep.java:40)
#31 289.0     	at org.gradle.internal.execution.steps.ChoosePipelineStep.execute(ChoosePipelineStep.java:23)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteWorkBuildOperationFiringStep.lambda$execute$2(ExecuteWorkBuildOperationFiringStep.java:67)
#31 289.0     	at java.base/java.util.Optional.orElseGet(Optional.java:364)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteWorkBuildOperationFiringStep.execute(ExecuteWorkBuildOperationFiringStep.java:67)
#31 289.0     	at org.gradle.internal.execution.steps.ExecuteWorkBuildOperationFiringStep.execute(ExecuteWorkBuildOperationFiringStep.java:39)
#31 289.0     	at org.gradle.internal.execution.steps.IdentityCacheStep.execute(IdentityCacheStep.java:46)
#31 289.0     	at org.gradle.internal.execution.steps.IdentityCacheStep.execute(IdentityCacheStep.java:34)
#31 289.0     	at org.gradle.internal.execution.steps.IdentifyStep.execute(IdentifyStep.java:48)
#31 289.0     	at org.gradle.internal.execution.steps.IdentifyStep.execute(IdentifyStep.java:35)
#31 289.0     	at org.gradle.internal.execution.impl.DefaultExecutionEngine$1.execute(DefaultExecutionEngine.java:61)
#31 289.0     	at org.gradle.api.internal.tasks.execution.ExecuteActionsTaskExecuter.executeIfValid(ExecuteActionsTaskExecuter.java:127)
#31 289.0     	at org.gradle.api.internal.tasks.execution.ExecuteActionsTaskExecuter.execute(ExecuteActionsTaskExecuter.java:116)
#31 289.0     	at org.gradle.api.internal.tasks.execution.FinalizePropertiesTaskExecuter.execute(FinalizePropertiesTaskExecuter.java:46)
#31 289.0     	at org.gradle.api.internal.tasks.execution.ResolveTaskExecutionModeExecuter.execute(ResolveTaskExecutionModeExecuter.java:51)
#31 289.0     	at org.gradle.api.internal.tasks.execution.SkipTaskWithNoActionsExecuter.execute(SkipTaskWithNoActionsExecuter.java:57)
#31 289.0     	at org.gradle.api.internal.tasks.execution.SkipOnlyIfTaskExecuter.execute(SkipOnlyIfTaskExecuter.java:74)
#31 289.0     	at org.gradle.api.internal.tasks.execution.CatchExceptionTaskExecuter.execute(CatchExceptionTaskExecuter.java:36)
#31 289.0     	at org.gradle.api.internal.tasks.execution.EventFiringTaskExecuter$1.executeTask(EventFiringTaskExecuter.java:77)
#31 289.0     	at org.gradle.api.internal.tasks.execution.EventFiringTaskExecuter$1.call(EventFiringTaskExecuter.java:55)
#31 289.0     	at org.gradle.api.internal.tasks.execution.EventFiringTaskExecuter$1.call(EventFiringTaskExecuter.java:52)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$CallableBuildOperationWorker.execute(DefaultBuildOperationRunner.java:209)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$CallableBuildOperationWorker.execute(DefaultBuildOperationRunner.java:204)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:66)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:166)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.call(DefaultBuildOperationRunner.java:53)
#31 289.0     	at org.gradle.api.internal.tasks.execution.EventFiringTaskExecuter.execute(EventFiringTaskExecuter.java:52)
#31 289.0     	at org.gradle.execution.plan.LocalTaskNodeExecutor.execute(LocalTaskNodeExecutor.java:42)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph$InvokeNodeExecutorsAction.execute(DefaultTaskExecutionGraph.java:331)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph$InvokeNodeExecutorsAction.execute(DefaultTaskExecutionGraph.java:318)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph$BuildOperationAwareExecutionAction.lambda$execute$0(DefaultTaskExecutionGraph.java:314)
#31 289.0     	at org.gradle.internal.operations.CurrentBuildOperationRef.with(CurrentBuildOperationRef.java:85)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph$BuildOperationAwareExecutionAction.execute(DefaultTaskExecutionGraph.java:314)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph$BuildOperationAwareExecutionAction.execute(DefaultTaskExecutionGraph.java:303)
#31 289.0     	at org.gradle.execution.plan.DefaultPlanExecutor$ExecutorWorker.execute(DefaultPlanExecutor.java:459)
#31 289.0     	at org.gradle.execution.plan.DefaultPlanExecutor$ExecutorWorker.run(DefaultPlanExecutor.java:376)
#31 289.0     	at org.gradle.execution.plan.DefaultPlanExecutor.process(DefaultPlanExecutor.java:111)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph.executeWithServices(DefaultTaskExecutionGraph.java:138)
#31 289.0     	at org.gradle.execution.taskgraph.DefaultTaskExecutionGraph.execute(DefaultTaskExecutionGraph.java:123)
#31 289.0     	at org.gradle.execution.SelectedTaskExecutionAction.execute(SelectedTaskExecutionAction.java:35)
#31 289.0     	at org.gradle.execution.DryRunBuildExecutionAction.execute(DryRunBuildExecutionAction.java:51)
#31 289.0     	at org.gradle.execution.BuildOperationFiringBuildWorkerExecutor$ExecuteTasks.call(BuildOperationFiringBuildWorkerExecutor.java:54)
#31 289.0     	at org.gradle.execution.BuildOperationFiringBuildWorkerExecutor$ExecuteTasks.call(BuildOperationFiringBuildWorkerExecutor.java:43)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$CallableBuildOperationWorker.execute(DefaultBuildOperationRunner.java:209)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$CallableBuildOperationWorker.execute(DefaultBuildOperationRunner.java:204)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:66)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner$2.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:166)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.execute(DefaultBuildOperationRunner.java:59)
#31 289.0     	at org.gradle.internal.operations.DefaultBuildOperationRunner.call(DefaultBuildOperationRunner.java:53)
#31 289.0     	at org.gradle.execution.BuildOperationFiringBuildWorkerExecutor.execute(BuildOperationFiringBuildWorkerExecutor.java:40)
#31 289.0     	at org.gradle.internal.build.DefaultBuildLifecycleController.lambda$executeTasks$10(DefaultBuildLifecycleController.java:313)
#31 289.0     	at org.gradle.internal.model.StateTransitionController.doTransition(StateTransitionController.java:266)
#31 289.0     	at org.gradle.internal.model.StateTransitionController.lambda$tryTransition$8(StateTransitionController.java:177)
#31 289.0     	at org.gradle.internal.work.DefaultSynchronizer.withLock(DefaultSynchronizer.java:44)
#31 289.0     	at org.gradle.internal.model.StateTransitionController.tryTransition(StateTransitionController.java:177)
#31 289.0     	at org.gradle.internal.build.DefaultBuildLifecycleController.executeTasks(DefaultBuildLifecycleController.java:304)
#31 289.0     	at org.gradle.internal.build.DefaultBuildWorkGraphController$DefaultBuildWorkGraph.runWork(DefaultBuildWorkGraphController.java:220)
#31 289.0     	at org.gradle.internal.work.DefaultWorkerLeaseService.withLocks(DefaultWorkerLeaseService.java:263)
#31 289.0     	at org.gradle.internal.work.DefaultWorkerLeaseService.runAsWorkerThread(DefaultWorkerLeaseService.java:127)
#31 289.0     	at org.gradle.composite.internal.DefaultBuildController.doRun(DefaultBuildController.java:181)
#31 289.0     	at org.gradle.composite.internal.DefaultBuildController.access$000(DefaultBuildController.java:50)
#31 289.0     	at org.gradle.composite.internal.DefaultBuildController$BuildOpRunnable.lambda$run$0(DefaultBuildController.java:198)
#31 289.0     	at org.gradle.internal.operations.CurrentBuildOperationRef.with(CurrentBuildOperationRef.java:85)
#31 289.0     	at org.gradle.composite.internal.DefaultBuildController$BuildOpRunnable.run(DefaultBuildController.java:198)
#31 289.0     	at java.base/java.util.concurrent.Executors$RunnableAdapter.call(Executors.java:539)
#31 289.0     	at java.base/java.util.concurrent.FutureTask.run(FutureTask.java:264)
#31 289.0     	at org.gradle.internal.concurrent.ExecutorPolicy$CatchAndRecordFailures.onExecute(ExecutorPolicy.java:64)
#31 289.0     	at org.gradle.internal.concurrent.AbstractManagedExecutor$1.run(AbstractManagedExecutor.java:48)
#31 289.0     	at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1136)
#31 289.0     	at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:635)
#31 289.0     	at java.base/java.lang.Thread.run(Thread.java:840)
#31 289.0     
#31 290.0 
#31 290.0 > Task :react-native-gesture-handler:mergeReleaseResources
#31 290.3 > Task :react-native-community_datetimepicker:mergeReleaseResources
#31 301.6 
#31 301.6 > Task :react-native-gesture-handler:compileReleaseKotlin
#31 301.6 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/core/FlingGestureHandler.kt:25:26 Parameter 'event' is never used
#31 301.6 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:79:62 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 301.6 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:84:63 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 301.6 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:89:65 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 301.6 w: file:///app/node_modules/react-native-gesture-handler/android/src/main/java/com/swmansion/gesturehandler/react/RNGestureHandlerButtonViewManager.kt:94:66 The corresponding parameter in the supertype 'ViewGroupManager' is named 'borderRadius'. This may cause problems when calling this function with named arguments.
#31 301.9 143 actionable tasks: 143 executed
#31 301.9 
#31 301.9 FAILURE: Build failed with an exception.
#31 301.9 
#31 301.9 * What went wrong:
#31 301.9 Execution failed for task ':app:mergeReleaseResources'.
#31 301.9 > /app/android/app/src/main/res/values/colors.xml:2:33: Error: The string "--" is not permitted within comments.
#31 301.9 
#31 301.9 * Try:
#31 301.9 > Run with --stacktrace option to get the stack trace.
#31 301.9 > Run with --info or --debug option to get more log output.
#31 301.9 > Run with --scan to get full insights.
#31 301.9 > Get more help at https://help.gradle.org.
#31 301.9 
#31 301.9 BUILD FAILED in 5m 1s
#31 ERROR: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
------
 > [mobile-build 8/8] RUN VERSION="$(cat VERSION)" &&     sed -i "s/versionName "[^"]*"/versionName "${VERSION}"/" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk":
301.9 Execution failed for task ':app:mergeReleaseResources'.
301.9 > /app/android/app/src/main/res/values/colors.xml:2:33: Error: The string "--" is not permitted within comments.
301.9 
301.9 * Try:
301.9 > Run with --stacktrace option to get the stack trace.
301.9 > Run with --info or --debug option to get more log output.
301.9 > Run with --scan to get full insights.
301.9 > Get more help at https://help.gradle.org.
301.9 
301.9 BUILD FAILED in 5m 1s
------
Dockerfile:59
--------------------
  58 |     COPY VERSION ./VERSION
  59 | >>> RUN VERSION="$(cat VERSION)" && \
  60 | >>>     sed -i "s/versionName \"[^\"]*\"/versionName \"${VERSION}\"/" android/app/build.gradle && \
  61 | >>>     cd android && gradle assembleRelease --no-daemon && \
  62 | >>>     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk "/out/cwclock-v${VERSION}.apk"
  63 |     
--------------------
ERROR: failed to solve: process "/bin/sh -c VERSION=\"$(cat VERSION)\" &&     sed -i \"s/versionName \\\"[^\\\"]*\\\"/versionName \\\"${VERSION}\\\"/\" android/app/build.gradle &&     cd android && gradle assembleRelease --no-daemon &&     mkdir -p /out && cp app/build/outputs/apk/release/app-release.apk \"/out/cwclock-v${VERSION}.apk\"" did not complete successfully: exit code: 1
```