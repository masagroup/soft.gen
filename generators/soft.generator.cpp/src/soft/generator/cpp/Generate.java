package soft.generator.cpp;

import java.io.IOException;
import java.net.URI;
import java.net.URL;
import java.nio.file.FileSystem;
import java.nio.file.FileSystemNotFoundException;
import java.nio.file.FileSystems;
import java.nio.file.FileVisitResult;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.SimpleFileVisitor;
import java.nio.file.StandardCopyOption;
import java.nio.file.attribute.BasicFileAttributes;
import java.util.Arrays;
import java.util.Collections;
import java.util.Map;

import org.eclipse.emf.common.util.BasicMonitor;
import org.eclipse.emf.ecore.EObject;
import org.eclipse.emf.ecore.EPackage;
import org.eclipse.emf.ecore.resource.ResourceSet;

import soft.generator.common.Generator;

/**
 * Entry point of the 'Generate' generation module.
 * 
 * @generated NOT
 */
public class Generate extends Generator {
    /**
     * The name of the module.
     */
    public static final String MODULE_FILE_NAME = "/soft/generator/cpp/generate";

    /**
     * The name of the templates that are to be generated.
     */
    public static final String[] TEMPLATE_NAMES = { "generateModel",
                                                    "generateLibraryCMakeSources",
                                                    "generateLibraryCMakeProject",
                                                    "generateTests",
                                                    "generateTestsCMakeSources",
                                                    "generateTestsCMakeProject" };

    private Generate() {
        super(MODULE_FILE_NAME, TEMPLATE_NAMES);
    }

    public static void main(String[] args) {
        try {
            Generate g = new Generate();
            boolean p = g.parse(args);
            if (p) {
                g.initialize();
                g.generate(new BasicMonitor());
            }
        } catch (Exception e) {
            System.out.println(e.getMessage());
        }
    }

    private static class PathReference implements AutoCloseable {

        private Path p;
        private FileSystem fs;

        public PathReference(Path p, FileSystem fs) {
            this.p = p;
            this.fs = fs;
        }

        public void close() throws Exception {
            if (this.fs != null)
                this.fs.close();
        }

        public Path getPath() {
            return p;
        }
    }

    private static PathReference getPath(final URI resURI) throws IOException {
        try {
            // first try getting a path via existing file systems
            return new PathReference(Paths.get(resURI), null);
        } catch (final FileSystemNotFoundException e) {
            /*
             * not directly on file system, so then it's somewhere else (e.g.: JAR)
             */
            final Map<String, ?> env = Collections.emptyMap();
            final FileSystem fs = FileSystems.newFileSystem(resURI, env);
            return new PathReference(fs.provider().getPath(resURI), fs);
        }
    }

    protected void postGenerate(ResourceSet resourceSet) {
        if (Arrays.stream(getTemplateNames()).anyMatch(t -> t.equals("generateTestsCMakeProject"))) {
            EObject eModel = getModel();
            if (eModel instanceof EPackage) {
                EPackage ePackage = (EPackage) eModel;
                try {
                    URL srcURL = getClass().getResource("cmake");
                    URI srcURI = srcURL.toURI();

                    // for each file in cmake src resource directory, copy it to destination
                    PathReference fromReference = getPath(srcURI);
                    PathReference toReference = getPath(targetFolder.toURI());
                    Path from = fromReference.getPath().normalize();
                    Path to = toReference.getPath().resolve(ePackage.getName() + "/tests/cmake").normalize();

                    Files.walkFileTree(from, new SimpleFileVisitor<Path>() {

                        @Override
                        public FileVisitResult preVisitDirectory(final Path dir, final BasicFileAttributes attrs)
                                throws IOException {
                            Path dest = to.resolve(from.relativize(dir).toString());
                            System.out.println("creating directory:" + dest);
                            Files.createDirectories(dest);
                            return FileVisitResult.CONTINUE;
                        }

                        @Override
                        public FileVisitResult visitFile(final Path file, final BasicFileAttributes attrs)
                                throws IOException {
                            Path dest = to.resolve(from.relativize(file).toString());
                            System.out.println("extracting file:" + dest);
                            Files.copy(file, dest, StandardCopyOption.REPLACE_EXISTING);
                            return FileVisitResult.CONTINUE;
                        }
                    });
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }

        }
    }

}
