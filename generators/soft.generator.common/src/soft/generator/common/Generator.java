package soft.generator.common;

import java.io.File;
import java.io.IOException;
import java.net.URL;
import java.nio.file.FileSystems;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Properties;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

import org.apache.commons.cli.CommandLine;
import org.apache.commons.cli.CommandLineParser;
import org.apache.commons.cli.DefaultParser;
import org.apache.commons.cli.HelpFormatter;
import org.apache.commons.cli.Option;
import org.apache.commons.cli.Options;
import org.apache.commons.cli.ParseException;
import org.eclipse.acceleo.common.IAcceleoConstants;
import org.eclipse.acceleo.common.internal.utils.AcceleoPackageRegistry;
import org.eclipse.acceleo.common.utils.ModelUtils;
import org.eclipse.acceleo.engine.event.AcceleoTextGenerationEvent;
import org.eclipse.acceleo.engine.event.IAcceleoTextGenerationListener;
import org.eclipse.acceleo.engine.service.AbstractAcceleoGenerator;
import org.eclipse.acceleo.engine.service.AcceleoService;
import org.eclipse.acceleo.model.mtl.Module;
import org.eclipse.acceleo.model.mtl.resource.AcceleoResourceFactoryRegistry;
import org.eclipse.acceleo.model.mtl.resource.AcceleoResourceSetImpl;
import org.eclipse.emf.common.util.URI;
import org.eclipse.emf.ecore.plugin.EcorePlugin;
import org.eclipse.emf.ecore.resource.Resource;
import org.eclipse.emf.ecore.resource.Resource.Factory;
import org.eclipse.emf.ecore.resource.ResourceSet;
import org.eclipse.emf.ecore.resource.URIConverter;
import org.eclipse.emf.ecore.resource.impl.ExtensibleURIConverterImpl;

import com.google.common.collect.Lists;
import com.google.common.collect.Maps;

/**
 * Entry point of the 'Generate' generation module.
 */
public class Generator extends AbstractAcceleoGenerator {

    private class ResourceFactoryRegistry extends AcceleoResourceFactoryRegistry {
        public ResourceFactoryRegistry(Resource.Factory.Registry delegate) {
            super(delegate);
        }

        protected Factory delegatedGetFactory(URI uri, String contentTypeIdentifier) {
            if ("classpath".equals(uri.scheme())) {
                URL url = getClass().getResource(uri.path());
                uri = createTemplateURI(url.toString());
            }
            return super.delegatedGetFactory(uri, contentTypeIdentifier);
        }
    }

    private String moduleName;
    private String[] defaultTemplates;

    private String modelPath;
    private String targetPath;
    private List<String> templates;
    private Properties properties = new Properties();
    private List<String> propertiesFiles = new ArrayList<String>();
    private boolean silentMode = false;

    protected Generator(String moduleName, String[] defaultTemplates) {
        this.moduleName = moduleName;
        this.defaultTemplates = defaultTemplates;
    }

    public boolean parse(String[] args) {
        Options generateOptions = new Options();
        Option helpOption = new Option("help", "print this message");
        Option templateOption = Option.builder("t")
                                      .argName("templates")
                                      .longOpt("templates")
                                      .required(false)
                                      .hasArgs()
                                      .desc("the templates to be executed : "
                                              + Arrays.stream(defaultTemplates).collect(Collectors.joining(", ")))
                                      .build();
        Option modelOption = Option.builder("m")
                                   .argName("model")
                                   .longOpt("model")
                                   .required()
                                   .hasArg()
                                   .desc("the input model")
                                   .build();
        Option outputOption = Option.builder("o")
                                    .argName("folder")
                                    .longOpt("output")
                                    .required()
                                    .hasArg()
                                    .desc("the output folder")
                                    .build();
        Option propertyOption = Option.builder("p")
                                      .argName("property=value")
                                      .desc("a property")
                                      .valueSeparator('=')
                                      .numberOfArgs(2)
                                      .build();
        Option silentOption = Option.builder("s")
                                    .longOpt("silent")
                                    .hasArg(false)
                                    .desc("print nothing but failures")
                                    .build();

        generateOptions.addOption(helpOption);
        generateOptions.addOption(templateOption);
        generateOptions.addOption(modelOption);
        generateOptions.addOption(outputOption);
        generateOptions.addOption(propertyOption);
        generateOptions.addOption(silentOption);

        HelpFormatter help = new HelpFormatter();
        CommandLineParser parser = new DefaultParser();
        try {
            CommandLine line = parser.parse(generateOptions, args);
            if (line.hasOption("help")) {
                help.printHelp("Generate", generateOptions);
                return false;
            }
            modelPath = line.getOptionValue("m");
            targetPath = line.getOptionValue("o");

            // templates
            templates = Lists.newArrayList(defaultTemplates);
            if (line.hasOption("t")) {
                String[] regexs = line.getOptionValues("t");
                for (String regex : regexs) {
                    // check if pattern is excluded
                    boolean exclude = false;
                    if (regex.charAt(0) == '!') {
                        exclude = true;
                        regex = regex.substring(1);
                    }

                    // check for all defined templates if they got
                    // to be excluded or not
                    Iterator<String> it = templates.iterator();
                    while (it.hasNext()) {
                        boolean match = Pattern.matches(regex, it.next());
                        if ((match && exclude) || (!match && !exclude))
                            it.remove();
                    }
                }
            }

            // properties
            properties = new Properties();
            if (line.hasOption("p"))
                properties = line.getOptionProperties("p");

            // silent mode
            silentMode = false;
            if (line.hasOption("s"))
                silentMode = true;
        } catch (ParseException e) {
            help.printHelp("Generate", generateOptions);
            return false;
        }
        return true;
    }

    public void initialize() throws IOException {
        ResourceSet modulesResourceSet = new AcceleoResourceSetImpl();
        modulesResourceSet.setPackageRegistry(AcceleoPackageRegistry.INSTANCE);

        resourceFactoryRegistry = modulesResourceSet.getResourceFactoryRegistry();
        modulesResourceSet.setResourceFactoryRegistry(new ResourceFactoryRegistry(resourceFactoryRegistry));

        URIConverter uriConverter = createURIConverter();
        if (uriConverter != null) {
            modulesResourceSet.setURIConverter(uriConverter);
        }

        Map<URI, URI> uriMap = EcorePlugin.computePlatformURIMap(false);

        // make sure that metamodel projects in the workspace override those in plugins
        modulesResourceSet.getURIConverter().getURIMap().putAll(uriMap);

        registerResourceFactories(modulesResourceSet);
        registerPackages(modulesResourceSet);

        ResourceSet modelResourceSet = new AcceleoResourceSetImpl();
        modelResourceSet.setPackageRegistry(AcceleoPackageRegistry.INSTANCE);
        if (uriConverter != null) {
            modelResourceSet.setURIConverter(uriConverter);
        }

        // make sure that metamodel projects in the workspace override those in plugins
        modelResourceSet.getURIConverter().getURIMap().putAll(uriMap);

        registerResourceFactories(modelResourceSet);
        registerPackages(modelResourceSet);

        String moduleName = getModuleName();
        if (moduleName.endsWith('.' + IAcceleoConstants.MTL_FILE_EXTENSION)) {
            moduleName = moduleName.substring(0, moduleName.lastIndexOf('.'));
        }
        if (!moduleName.endsWith('.' + IAcceleoConstants.EMTL_FILE_EXTENSION)) {
            moduleName += '.' + IAcceleoConstants.EMTL_FILE_EXTENSION;
        }

        URL moduleURL = findModuleURL(moduleName);

        if (moduleURL == null) {
            throw new IOException("'" + getModuleName() + ".emtl' not found"); //$NON-NLS-1$ //$NON-NLS-2$
        }
        URI moduleURI = createTemplateURI(moduleURL.toString());
        moduleURI = URI.createURI(moduleURI.toString(), true);
        module = (Module) ModelUtils.load(moduleURI, modulesResourceSet);

        URI newModelURI = URI.createURI(modelPath, true);
        model = ModelUtils.load(newModelURI, modelResourceSet);
        targetFolder = new File(targetPath);
        generationArguments = Collections.emptyList();
        this.postInitialize();
    }

    protected AcceleoService createAcceleoService() {
        AcceleoService service = super.createAcceleoService();
        service.addProperties(Maps.fromProperties(properties));
        return service;
    }

    /**
     * If this generator needs to listen to text generation events, listeners can be
     * returned from here.
     * 
     * @return List of listeners that are to be notified when text is generated
     *         through this launch.
     * @generated NOT
     */
    @Override
    public List<IAcceleoTextGenerationListener> getGenerationListeners() {
        List<IAcceleoTextGenerationListener> listeners = super.getGenerationListeners();
        if (!silentMode)
            listeners.add(new IAcceleoTextGenerationListener() {

                @Override
                public void textGenerated(AcceleoTextGenerationEvent event) {
                }

                @Override
                public boolean listensToGenerationEnd() {
                    return false;
                }

                @Override
                public void generationEnd(AcceleoTextGenerationEvent event) {
                }

                @Override
                public void filePathComputed(AcceleoTextGenerationEvent event) {
                }

                @Override
                public void fileGenerated(AcceleoTextGenerationEvent event) {
                    Path path = FileSystems.getDefault().getPath((String) event.getText());
                    System.out.println("file generated: " + path.normalize());
                }
            });
        return listeners;
    }

    /**
     * This will be called in order to find and load the module that will be
     * launched through this launcher. We expect this name not to contain file
     * extension, and the module to be located beside the launcher.
     * 
     * @return The name of the module that is to be launched.
     * @generated
     */
    @Override
    public String getModuleName() {
        return moduleName;
    }

    /**
     * If the module(s) called by this launcher require properties files, return
     * their qualified path from here.Take note that the first added properties
     * files will take precedence over subsequent ones if they contain conflicting
     * keys.
     * 
     * @return The list of properties file we need to add to the generation context.
     * @see java.util.ResourceBundle#getBundle(String)
     * @generated
     */
    @Override
    public List<String> getProperties() {
        return propertiesFiles;
    }

    /**
     * Adds a properties file in the list of properties files.
     * 
     * @param propertiesFile The properties file to add.
     * @generated
     * @since 3.1
     */
    @Override
    public void addPropertiesFile(String propertiesFile) {
        this.propertiesFiles.add(propertiesFile);
    }

    /**
     * This will be used to get the list of templates that are to be launched by
     * this launcher.
     * 
     * @return The list of templates to call on the module {@link #getModuleName()}.
     * @generated NOT
     */
    @Override
    public String[] getTemplateNames() {
        return (String[]) templates.toArray(new String[templates.size()]);
    }

    /**
     * This can be used to update the resource set's package registry with all
     * needed EPackages.
     * 
     * @param resourceSet The resource set which registry has to be updated.
     * @generated
     */
    @Override
    public void registerPackages(ResourceSet resourceSet) {
        super.registerPackages(resourceSet);
        if (!isInWorkspace(org.eclipse.emf.ecore.EcorePackage.class)) {
            resourceSet.getPackageRegistry()
                       .put(org.eclipse.emf.ecore.EcorePackage.eINSTANCE.getNsURI(),
                            org.eclipse.emf.ecore.EcorePackage.eINSTANCE);
        }
    }

    protected URIConverter createURIConverter() {
        return new ExtensibleURIConverterImpl() {
            /**
             * {@inheritDoc}
             * 
             * @see org.eclipse.emf.ecore.resource.impl.ExtensibleURIConverterImpl#normalize(org.eclipse.emf.common.util.URI)
             */
            @Override
            public URI normalize(URI uri) {
                URI normalized = getURIMap().get(uri);
                if (normalized == null) {
                    if ("classpath".equals(uri.scheme())) {
                        URL url = getClass().getResource(uri.path());
                        normalized = createTemplateURI(url.toString());
                        getURIMap().put(uri, normalized);
                    }
                }
                return normalized != null ? normalized : super.normalize(uri);
            }
        };
    }

}
