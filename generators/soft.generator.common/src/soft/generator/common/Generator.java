package soft.generator.common;

import java.io.File;
import java.io.IOException;
import java.nio.file.FileSystems;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.Iterator;
import java.util.List;
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
import org.eclipse.acceleo.engine.event.AcceleoTextGenerationEvent;
import org.eclipse.acceleo.engine.event.IAcceleoTextGenerationListener;
import org.eclipse.acceleo.engine.service.AbstractAcceleoGenerator;
import org.eclipse.acceleo.engine.service.AcceleoService;
import org.eclipse.emf.common.util.BasicMonitor;
import org.eclipse.emf.common.util.URI;
import org.eclipse.emf.ecore.EObject;
import org.eclipse.emf.ecore.resource.ResourceSet;

import com.google.common.collect.Lists;
import com.google.common.collect.Maps;

/**
 * Entry point of the 'Generate' generation module.
 */
public class Generator extends AbstractAcceleoGenerator {

    /**
     * The list of properties files from the launch parameters (Launch
     * configuration).
     */
    private String moduleName;
    private List<String> templates;
    private Properties properties = new Properties();
    private List<String> propertiesFiles = new ArrayList<String>();
    private boolean silent = false;

    protected Generator() {

    }

    /**
     * This allows clients to instantiates a generator with all required
     * information.
     * 
     * @param modelURI     URI where the model on which this generator will be used
     *                     is located.
     * @param targetFolder This will be used as the output folder for this
     *                     generation : it will be the base path against which all
     *                     file block URLs will be resolved.
     * @param arguments    If the template which will be called requires more than
     *                     one argument taken from the model, pass them here.
     * @throws IOException This can be thrown in three scenarios : the module cannot
     *                     be found, it cannot be loaded, or the model cannot be
     *                     loaded.
     * @generated
     */
    private Generator(URI modelURI, File targetFolder, List<? extends Object> arguments) throws IOException {
        initialize(modelURI, targetFolder, arguments);
    }

    public void initialize(URI modelURI, File targetFolder, List<? extends Object> arguments) throws IOException {
        initializeArguments(arguments);
        super.initialize(modelURI, targetFolder, Collections.emptyList());
    }

    public void initialize(EObject model, File targetFolder, List<? extends Object> arguments) throws IOException {
        initializeArguments(arguments);
        super.initialize(model, targetFolder, Collections.emptyList());
    }

    private void initializeArguments(List<? extends Object> arguments) {
        this.moduleName = (String) arguments.get(0);
        @SuppressWarnings("unchecked")
        List<String> templates = (List<String>) arguments.get(1);
        this.templates = templates;
        Properties properties = (Properties) arguments.get(2);
        this.properties = properties;
        Boolean silent = (Boolean) arguments.get(3);
        this.silent = silent;
    }

    protected AcceleoService createAcceleoService() {
        AcceleoService service = super.createAcceleoService();
        service.addProperties(Maps.fromProperties(properties));
        return service;
    }

    /**
     * This can be used to launch the generation from a standalone application.
     * 
     * @param args Arguments of the generation.
     */
    public static void generate(String[] args, String moduleName, String[] defaultTemplates) {
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
                return;
            }
            URI model = URI.createFileURI(line.getOptionValue("m"));
            File output = new File(line.getOptionValue("o"));

            // templates
            List<String> templates = Lists.newArrayList(defaultTemplates);
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
            Properties properties = new Properties();
            if (line.hasOption("p"))
                properties = line.getOptionProperties("p");

            // silent mode
            boolean silentMode = false;
            if (line.hasOption("s"))
                silentMode = true;

            Generator generator = new Generator(model,
                                                output,
                                                Arrays.asList(moduleName, templates, properties, silentMode));
            generator.doGenerate(new BasicMonitor());

        } catch (ParseException e) {
            System.out.println(e.getMessage());
            help.printHelp("Generate", generateOptions);
        } catch (IOException e) {
            e.printStackTrace();
        }
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
        if (!silent)
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

}
